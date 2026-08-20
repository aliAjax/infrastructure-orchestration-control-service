package application

import (
	"encoding/json"
	"fmt"

	"github.com/infra-orchestration/controlplane/internal/execution/domain"
)

type Checkpoint struct {
	ResourceID string          `json:"resource_id"`
	Operation  string          `json:"operation"`
	InputHash  string          `json:"input_hash"`
	Output     json.RawMessage `json:"output"`
}

func BuildCheckpoint(task domain.Task, output json.RawMessage) Checkpoint {
	return Checkpoint{
		ResourceID: task.ResourceID,
		Operation:  string(task.Status),
		InputHash:  hashRaw(task.Input),
		Output:     output,
	}
}

func (c Checkpoint) Validate() error {
	if c.ResourceID == "" || c.Output == nil {
		return fmt.Errorf("checkpoint requires resource_id and output")
	}
	return nil
}

func hashRaw(raw json.RawMessage) string {
	var value any
	_ = json.Unmarshal(raw, &value)
	data, _ := json.Marshal(value)
	return string(data)
}
