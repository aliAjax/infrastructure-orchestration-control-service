package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/infra-orchestration/controlplane/internal/execution/domain"
)

type MockRunner struct{}

func NewMockRunner() *MockRunner {
	return &MockRunner{}
}

func (r *MockRunner) Execute(ctx context.Context, request domain.RunnerRequest) (domain.RunnerResponse, error) {
	select {
	case <-ctx.Done():
		return domain.RunnerResponse{}, ctx.Err()
	case <-time.After(20 * time.Millisecond):
	}
	output, _ := json.Marshal(map[string]any{
		"resource_id": request.ResourceID,
		"type":        request.Type,
		"result":      "ok",
		"ts":          time.Now().UTC(),
	})
	return domain.RunnerResponse{Output: output}, nil
}

func (r *MockRunner) String() string {
	return fmt.Sprintf("mock-runner")
}
