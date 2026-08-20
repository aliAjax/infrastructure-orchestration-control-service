package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	server := flag.String("server", "http://localhost:8080", "control plane HTTP endpoint")
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}
	client := &apiClient{base: strings.TrimRight(*server, "/")}
	var err error
	switch args[0] {
	case "health":
		err = client.health()
	case "projects":
		err = client.projects(args[1:])
	case "environments":
		err = client.environments(args[1:])
	case "resources":
		err = client.resources(args[1:])
	case "plans":
		err = client.plans(args[1:])
	case "approvals":
		err = client.approvals(args[1:])
	case "executions":
		err = client.executions(args[1:])
	case "drift":
		err = client.drift(args[1:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`usage: infra-cli -server URL <command> [args]
commands:
  health
  projects list
  projects create <name> [description]
  environments list [project_id]
  environments create <project_id> <name> <kind> <production>
  resources list <environment_id>
  resources upsert <environment_id> <name> <type> <desired_json>
  plans generate <environment_id> [created_by]
  plans list <environment_id>
  approvals create <plan_id> <environment_id> <requested_by> <reason>
  approvals approve <approval_id> <approved_by>
  executions start <plan_id>
  executions list <plan_id>
  drift detect <environment_id>
  drift list <environment_id>`)
}

type apiClient struct {
	base string
}

func (c *apiClient) health() error {
	return c.get("/healthz", nil)
}

func (c *apiClient) projects(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing projects subcommand")
	}
	switch args[0] {
	case "list":
		return c.get("/api/v1/projects", nil)
	case "create":
		if len(args) < 2 {
			return fmt.Errorf("usage: projects create <name> [description]")
		}
		desc := ""
		if len(args) > 2 {
			desc = args[2]
		}
		return c.post("/api/v1/projects", map[string]string{"name": args[1], "description": desc}, nil)
	default:
		return fmt.Errorf("unknown projects command %q", args[0])
	}
}

func (c *apiClient) environments(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing environments subcommand")
	}
	switch args[0] {
	case "list":
		projectID := ""
		if len(args) > 1 {
			projectID = args[1]
		}
		return c.get("/api/v1/environments?project_id="+projectID, nil)
	case "create":
		if len(args) < 5 {
			return fmt.Errorf("usage: environments create <project_id> <name> <kind> <production>")
		}
		production := args[4] == "true"
		return c.post("/api/v1/environments", map[string]any{"project_id": args[1], "name": args[2], "kind": args[3], "production": production}, nil)
	default:
		return fmt.Errorf("unknown environments command %q", args[0])
	}
}

func (c *apiClient) resources(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing resources subcommand")
	}
	switch args[0] {
	case "list":
		if len(args) < 2 {
			return fmt.Errorf("usage: resources list <environment_id>")
		}
		return c.get("/api/v1/resources?environment_id="+args[1], nil)
	case "upsert":
		if len(args) < 5 {
			return fmt.Errorf("usage: resources upsert <environment_id> <name> <type> <desired_json>")
		}
		var desired json.RawMessage
		if err := json.Unmarshal([]byte(args[4]), &desired); err != nil {
			return fmt.Errorf("invalid desired json: %w", err)
		}
		return c.post("/api/v1/resources", map[string]any{"environment_id": args[1], "name": args[2], "type": args[3], "provider": "mock", "desired_state": desired}, nil)
	default:
		return fmt.Errorf("unknown resources command %q", args[0])
	}
}

func (c *apiClient) plans(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing plans subcommand")
	}
	switch args[0] {
	case "generate":
		if len(args) < 2 {
			return fmt.Errorf("usage: plans generate <environment_id> [created_by]")
		}
		actor := "cli"
		if len(args) > 2 {
			actor = args[2]
		}
		return c.post("/api/v1/plans/generate", map[string]string{"environment_id": args[1], "created_by": actor}, nil)
	case "list":
		if len(args) < 2 {
			return fmt.Errorf("usage: plans list <environment_id>")
		}
		return c.get("/api/v1/plans?environment_id="+args[1], nil)
	default:
		return fmt.Errorf("unknown plans command %q", args[0])
	}
}

func (c *apiClient) approvals(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing approvals subcommand")
	}
	switch args[0] {
	case "create":
		if len(args) < 5 {
			return fmt.Errorf("usage: approvals create <plan_id> <environment_id> <requested_by> <reason>")
		}
		return c.post("/api/v1/approvals", map[string]string{"plan_id": args[1], "environment_id": args[2], "requested_by": args[3], "reason": args[4]}, nil)
	case "approve":
		if len(args) < 3 {
			return fmt.Errorf("usage: approvals approve <approval_id> <approved_by>")
		}
		return c.post("/api/v1/approvals/"+args[1]+"/decision", map[string]any{"approved_by": args[2], "approved": true, "decision_note": "approved from cli"}, nil)
	default:
		return fmt.Errorf("unknown approvals command %q", args[0])
	}
}

func (c *apiClient) executions(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing executions subcommand")
	}
	switch args[0] {
	case "start":
		if len(args) < 2 {
			return fmt.Errorf("usage: executions start <plan_id>")
		}
		return c.post("/api/v1/executions/start", map[string]any{"plan_id": args[1], "max_attempts": 3}, nil)
	case "list":
		if len(args) < 2 {
			return fmt.Errorf("usage: executions list <plan_id>")
		}
		return c.get("/api/v1/executions?plan_id="+args[1], nil)
	default:
		return fmt.Errorf("unknown executions command %q", args[0])
	}
}

func (c *apiClient) drift(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing drift subcommand")
	}
	switch args[0] {
	case "detect":
		if len(args) < 2 {
			return fmt.Errorf("usage: drift detect <environment_id>")
		}
		return c.post("/api/v1/drift/detect", map[string]string{"environment_id": args[1]}, nil)
	case "list":
		if len(args) < 2 {
			return fmt.Errorf("usage: drift list <environment_id>")
		}
		return c.get("/api/v1/drift?environment_id="+args[1], nil)
	default:
		return fmt.Errorf("unknown drift command %q", args[0])
	}
}

func (c *apiClient) get(path string, out any) error {
	resp, err := http.Get(c.base + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeResponse(resp, out)
}

func (c *apiClient) post(path string, body any, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := http.Post(c.base+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeResponse(resp, out)
}

func decodeResponse(resp *http.Response, out any) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}
	if out == nil {
		raw, _ := io.ReadAll(resp.Body)
		fmt.Println(strings.TrimSpace(string(raw)))
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
