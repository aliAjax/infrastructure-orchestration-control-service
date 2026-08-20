package adapter

import "github.com/infra-orchestration/controlplane/internal/resource/domain"

type ProjectDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func ProjectFromDomain(p domain.Project) ProjectDTO {
	return ProjectDTO{ID: p.ID, Name: p.Name, Description: p.Description}
}

type EnvironmentDTO struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Production bool   `json:"production"`
	LockKey    string `json:"lock_key"`
}

func EnvironmentFromDomain(e domain.Environment) EnvironmentDTO {
	return EnvironmentDTO{ID: e.ID, ProjectID: e.ProjectID, Name: e.Name, Kind: e.Kind, Production: e.Production, LockKey: e.LockKey}
}
