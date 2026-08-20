package application

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/infra-orchestration/controlplane/internal/resource/domain"
	"gopkg.in/yaml.v3"
)

type Declaration struct {
	EnvironmentID string                 `json:"environment_id" yaml:"environment_id"`
	Resources     []domain.ResourceInput `json:"resources" yaml:"resources"`
}

func ParseDeclaration(data []byte, format string) (Declaration, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	switch format {
	case "yaml", "yml":
		return parseYAML(data)
	case "json":
		return parseJSON(data)
	case "hcl":
		return parseHCL(data)
	default:
		return Declaration{}, fmt.Errorf("unsupported declaration format %q", format)
	}
}

func parseJSON(data []byte) (Declaration, error) {
	var decl Declaration
	if err := json.Unmarshal(data, &decl); err != nil {
		return Declaration{}, fmt.Errorf("parse JSON declaration: %w", err)
	}
	return validateDeclaration(decl)
}

func parseYAML(data []byte) (Declaration, error) {
	var decl Declaration
	if err := yaml.Unmarshal(data, &decl); err != nil {
		return Declaration{}, fmt.Errorf("parse YAML declaration: %w", err)
	}
	return validateDeclaration(decl)
}

func validateDeclaration(decl Declaration) (Declaration, error) {
	if decl.EnvironmentID == "" {
		return Declaration{}, fmt.Errorf("environment_id is required")
	}
	names := make(map[string]struct{}, len(decl.Resources))
	for i := range decl.Resources {
		res := &decl.Resources[i]
		res.Name = strings.TrimSpace(res.Name)
		res.Type = strings.TrimSpace(res.Type)
		if res.Name == "" || res.Type == "" {
			return Declaration{}, fmt.Errorf("resource at index %d requires name and type", i)
		}
		if _, exists := names[res.Name]; exists {
			return Declaration{}, fmt.Errorf("duplicate resource name %q", res.Name)
		}
		names[res.Name] = struct{}{}
		if len(res.DesiredState) == 0 {
			res.DesiredState = json.RawMessage(`{}`)
		}
		res.EnvironmentID = decl.EnvironmentID
		res.Provider = "mock"
		if res.LockKey == "" {
			res.LockKey = "resource:" + decl.EnvironmentID + ":" + strings.ToLower(res.Name)
		}
	}
	return decl, nil
}

// parseHCL implements a deliberately small HCL subset suitable for resource
// declarations. It is not a general HCL implementation, but covers the
// block/attribute shape used by this control plane.
func parseHCL(data []byte) (Declaration, error) {
	text := string(data)
	var envID string
	var resources []domain.ResourceInput
	for _, block := range splitHCLBlocks(text) {
		kind, labels, body, ok := parseHCLBlock(block)
		if !ok {
			continue
		}
		switch kind {
		case "environment":
			if len(labels) > 0 {
				envID = labels[0]
			}
		case "resource":
			if len(labels) < 2 {
				return Declaration{}, fmt.Errorf("resource block requires type and name labels")
			}
			attrs := parseHCLAttributes(body)
			name := labels[1]
			desired := parseHCLObject(attrs["state"])
			raw, _ := json.Marshal(desired)
			resources = append(resources, domain.ResourceInput{
				Name:         name,
				Type:         labels[0],
				DesiredState: raw,
				DependsOn:    parseStringList(attrs["depends_on"]),
				Sensitive:    strings.EqualFold(attrs["sensitive"], "true"),
			})
		}
	}
	return validateDeclaration(Declaration{EnvironmentID: envID, Resources: resources})
}

func splitHCLBlocks(text string) []string {
	var out []string
	start := -1
	depth := 0
	for i, r := range text {
		if r == '{' {
			if depth == 0 {
				start = i
			}
			depth++
		}
		if r == '}' {
			depth--
			if depth == 0 && start >= 0 {
				out = append(out, text[start:i+1])
				start = -1
			}
		}
	}
	return out
}

func parseHCLBlock(block string) (kind string, labels []string, body string, ok bool) {
	open := strings.IndexByte(block, '{')
	if open < 0 {
		return "", nil, "", false
	}
	header := strings.TrimSpace(block[:open])
	fields := strings.Fields(header)
	if len(fields) == 0 {
		return "", nil, "", false
	}
	kind = fields[0]
	for _, field := range fields[1:] {
		field = strings.Trim(field, `"`)
		labels = append(labels, field)
	}
	body = strings.TrimSpace(block[open+1 : len(block)-1])
	return kind, labels, body, true
}

func parseHCLAttributes(body string) map[string]string {
	attrs := make(map[string]string)
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.IndexByte(line, '='); idx >= 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			value = strings.Trim(value, `"`)
			attrs[key] = value
		}
	}
	return attrs
}

func parseHCLObject(text string) map[string]any {
	text = strings.TrimSpace(text)
	if text == "" {
		return map[string]any{}
	}
	var raw json.RawMessage
	if strings.HasPrefix(text, "{") {
		raw = json.RawMessage(text)
	} else {
		raw = json.RawMessage(`{"value":` + text + `}`)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{"value": strings.Trim(text, `"`)}
	}
	return out
}

func parseStringList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(strings.TrimSpace(part), `"`)
		if part != "" {
			out = append(out, part)
		}
	}
	sort.Strings(out)
	return out
}
