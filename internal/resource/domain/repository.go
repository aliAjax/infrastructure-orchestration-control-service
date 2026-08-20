package domain

import "context"

type Repository interface {
	CreateProject(ctx context.Context, input ProjectInput) (Project, error)
	GetProject(ctx context.Context, id string) (Project, error)
	ListProjects(ctx context.Context) ([]Project, error)
	UpdateProject(ctx context.Context, id string, input ProjectInput) (Project, error)

	CreateEnvironment(ctx context.Context, input EnvironmentInput) (Environment, error)
	GetEnvironment(ctx context.Context, id string) (Environment, error)
	ListEnvironments(ctx context.Context, projectID string) ([]Environment, error)

	UpsertResource(ctx context.Context, input ResourceInput) (Resource, error)
	GetResource(ctx context.Context, id string) (Resource, error)
	ListResources(ctx context.Context, environmentID string) ([]Resource, error)

	UpsertVariable(ctx context.Context, input VariableInput) (Variable, error)
	GetVariable(ctx context.Context, projectID, key string) (Variable, error)
	ListVariables(ctx context.Context, projectID string) ([]Variable, error)
}
