package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/infra-orchestration/controlplane/internal/audit/domain"
)

func (s *Service) RecordPlanCreated(ctx context.Context, actor, planID string, payload any) (domain.Event, error) {
	content, _ := json.Marshal(payload)
	return s.Record(ctx, domain.CreateInput{
		Type:       domain.EventPlanCreated,
		Actor:      actor,
		EntityType: "plan",
		EntityID:   planID,
		Content:    content,
		Result:     "created",
	})
}

func (s *Service) RecordApproval(ctx context.Context, actor, requestID string, approved bool) (domain.Event, error) {
	result := "approved"
	if !approved {
		result = "rejected"
	}
	return s.Record(ctx, domain.CreateInput{
		Type:       domain.EventPlanApproved,
		Actor:      actor,
		EntityType: "approval",
		EntityID:   requestID,
		Content:    json.RawMessage(fmt.Sprintf(`{"approved":%t}`, approved)),
		Result:     result,
	})
}

func (s *Service) RecordExecution(ctx context.Context, actor, taskID string, status string, err error) (domain.Event, error) {
	eventType := domain.EventExecutionSucceeded
	message := ""
	if err != nil {
		eventType = domain.EventExecutionFailed
		message = err.Error()
	}
	return s.Record(ctx, domain.CreateInput{
		Type:       eventType,
		Actor:      actor,
		EntityType: "execution_task",
		EntityID:   taskID,
		Content:    json.RawMessage(fmt.Sprintf(`{"status":%q}`, status)),
		Result:     status,
		Error:      message,
	})
}
