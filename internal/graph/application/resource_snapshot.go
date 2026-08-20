package application

import "github.com/infra-orchestration/controlplane/internal/resource/domain"

func cloneResourceSnapshot(resources []domain.Resource) []domain.Resource {
	if resources == nil {
		return nil
	}
	out := make([]domain.Resource, len(resources))
	copy(out, resources)
	for i := range out {
		out[i].DependsOn = append([]string(nil), resources[i].DependsOn...)
		out[i].DesiredState = append([]byte(nil), resources[i].DesiredState...)
	}
	return out
}
