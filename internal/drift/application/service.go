package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	driftdomain "github.com/infra-orchestration/controlplane/internal/drift/domain"
	planapplication "github.com/infra-orchestration/controlplane/internal/plan/application"
	plandomain "github.com/infra-orchestration/controlplane/internal/plan/domain"
	resourcedomain "github.com/infra-orchestration/controlplane/internal/resource/domain"
	statedomain "github.com/infra-orchestration/controlplane/internal/state/domain"
)

type Service struct {
	repo         driftdomain.Repository
	resourceRepo resourceRepo
	stateRepo    stateRepo
	planService  *planapplication.Service
}

type resourceRepo interface {
	ListResources(ctx context.Context, environmentID string) ([]resourcedomain.Resource, error)
}

type stateRepo interface {
	ListByEnvironment(ctx context.Context, environmentID string) ([]statedomain.ResourceState, error)
}

func NewService(repo driftdomain.Repository, resources resourceRepo, states stateRepo, plans *planapplication.Service) *Service {
	return &Service{repo: repo, resourceRepo: resources, stateRepo: states, planService: plans}
}

func (s *Service) Detect(ctx context.Context, environmentID string) ([]driftdomain.DriftRecord, error) {
	resources, err := s.resourceRepo.ListResources(ctx, environmentID)
	if err != nil {
		return nil, fmt.Errorf("list resources: %w", err)
	}
	states, err := s.stateRepo.ListByEnvironment(ctx, environmentID)
	if err != nil {
		return nil, fmt.Errorf("list states: %w", err)
	}
	stateMap := make(map[string]statedomain.ResourceState, len(states))
	for _, st := range states {
		stateMap[st.ResourceID] = st
	}
	var out []driftdomain.DriftRecord
	for _, res := range resources {
		st, ok := stateMap[res.ID]
		if !ok || len(st.ActualState) == 0 {
			continue
		}
		if !jsonEqual(res.DesiredState, st.ActualState) {
			record, err := s.repo.Create(ctx, driftdomain.CreateInput{
				EnvironmentID: environmentID,
				ResourceID:    res.ID,
				DesiredState:  res.DesiredState,
				ActualState:   st.ActualState,
				Severity:      "high",
			})
			if err != nil {
				return nil, err
			}
			out = append(out, record)
		}
	}
	return out, nil
}

func (s *Service) List(ctx context.Context, environmentID string, unresolvedOnly bool) ([]driftdomain.DriftRecord, error) {
	return s.repo.ListByEnvironment(ctx, environmentID, unresolvedOnly)
}

func (s *Service) Remediate(ctx context.Context, environmentID string) (plandomain.Plan, error) {
	records, err := s.repo.ListByEnvironment(ctx, environmentID, true)
	if err != nil {
		return plandomain.Plan{}, err
	}
	if len(records) == 0 {
		return plandomain.Plan{}, fmt.Errorf("no unresolved drift")
	}
	plan, err := s.planService.Generate(ctx, plandomain.GenerateInput{EnvironmentID: environmentID, CreatedBy: "drift-detector"})
	if err != nil {
		return plandomain.Plan{}, err
	}
	for _, record := range records {
		if _, err := s.repo.Resolve(ctx, driftdomain.ResolveInput{ID: record.ID, RemedyPlanID: plan.ID}); err != nil {
			return plandomain.Plan{}, err
		}
	}
	return plan, nil
}

func jsonEqual(a, b json.RawMessage) bool {
	var av, bv any
	_ = json.Unmarshal(a, &av)
	_ = json.Unmarshal(b, &bv)
	ab, _ := json.Marshal(av)
	bb, _ := json.Marshal(bv)
	return bytes.Equal(ab, bb)
}
