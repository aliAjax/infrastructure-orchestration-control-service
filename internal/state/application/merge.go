package application

import (
	"encoding/json"
	"fmt"
	"sort"
)

func MergeDesiredState(previous, next json.RawMessage) (json.RawMessage, error) {
	if len(previous) == 0 {
		return next, nil
	}
	if len(next) == 0 {
		return previous, nil
	}
	var a map[string]any
	var b map[string]any
	if err := json.Unmarshal(previous, &a); err != nil {
		return nil, fmt.Errorf("unmarshal previous state: %w", err)
	}
	if err := json.Unmarshal(next, &b); err != nil {
		return nil, fmt.Errorf("unmarshal next state: %w", err)
	}
	merged := deepMerge(a, b)
	out, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("marshal merged state: %w", err)
	}
	return out, nil
}

func deepMerge(base, override map[string]any) map[string]any {
	out := make(map[string]any, len(base)+len(override))
	for key, value := range base {
		out[key] = value
	}
	for key, value := range override {
		if nextMap, ok := value.(map[string]any); ok {
			if prevMap, exists := out[key].(map[string]any); exists {
				out[key] = deepMerge(prevMap, nextMap)
				continue
			}
		}
		out[key] = value
	}
	return out
}

func StateKeys(state json.RawMessage) []string {
	var obj map[string]any
	if err := json.Unmarshal(state, &obj); err != nil {
		return nil
	}
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
