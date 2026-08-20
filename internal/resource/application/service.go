package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/infra-orchestration/controlplane/internal/resource/domain"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateProject(ctx context.Context, in domain.ProjectInput) (domain.Project, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return domain.Project{}, fmt.Errorf("project name is required")
	}
	return s.repo.CreateProject(ctx, in)
}

func (s *Service) GetProject(ctx context.Context, id string) (domain.Project, error) {
	return s.repo.GetProject(ctx, id)
}

func (s *Service) ListProjects(ctx context.Context) ([]domain.Project, error) {
	return s.repo.ListProjects(ctx)
}

func (s *Service) UpdateProject(ctx context.Context, id string, input domain.ProjectInput) (domain.Project, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return domain.Project{}, fmt.Errorf("project name is required")
	}
	return s.repo.UpdateProject(ctx, id, input)
}

func (s *Service) CreateEnvironment(ctx context.Context, in domain.EnvironmentInput) (domain.Environment, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return domain.Environment{}, fmt.Errorf("environment name is required")
	}
	if in.ProjectID == "" {
		return domain.Environment{}, fmt.Errorf("project_id is required")
	}
	if in.LockKey == "" {
		in.LockKey = "env:" + in.ProjectID + ":" + strings.ToLower(in.Name)
	}
	return s.repo.CreateEnvironment(ctx, in)
}

func (s *Service) ListEnvironments(ctx context.Context, projectID string) ([]domain.Environment, error) {
	return s.repo.ListEnvironments(ctx, projectID)
}

func (s *Service) GetEnvironment(ctx context.Context, id string) (domain.Environment, error) {
	return s.repo.GetEnvironment(ctx, id)
}

func (s *Service) UpsertResource(ctx context.Context, in domain.ResourceInput) (domain.Resource, error) {
	if in.Name == "" || in.Type == "" {
		return domain.Resource{}, fmt.Errorf("resource name and type are required")
	}
	if in.Provider == "" {
		in.Provider = "mock"
	}
	if len(in.DesiredState) == 0 || string(in.DesiredState) == "null" {
		return domain.Resource{}, fmt.Errorf("desired_state is required")
	}
	if in.LockKey == "" {
		in.LockKey = "resource:" + in.EnvironmentID + ":" + strings.ToLower(in.Name)
	}
	return s.repo.UpsertResource(ctx, in)
}

func (s *Service) GetResource(ctx context.Context, id string) (domain.Resource, error) {
	return s.repo.GetResource(ctx, id)
}

func (s *Service) ListResources(ctx context.Context, environmentID string) ([]domain.Resource, error) {
	return s.repo.ListResources(ctx, environmentID)
}

func (s *Service) UpsertVariable(ctx context.Context, in domain.VariableInput) (domain.Variable, error) {
	if in.Key == "" || in.ProjectID == "" {
		return domain.Variable{}, fmt.Errorf("project_id and key are required")
	}
	return s.repo.UpsertVariable(ctx, in)
}

func (s *Service) ListVariables(ctx context.Context, projectID string) ([]domain.Variable, error) {
	return s.repo.ListVariables(ctx, projectID)
}
