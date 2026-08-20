package application

import "github.com/infra-orchestration/controlplane/internal/resource/domain"

// FilterResources returns a detached list so callers can safely retain the input.
func FilterResources(resources []domain.Resource, keep func(domain.Resource) bool) []domain.Resource {
	out := make([]domain.Resource, 0, len(resources))
	for _, resource := range resources {
		if keep == nil || keep(resource) {
			out = append(out, cloneFilteredResource(resource))
		}
	}
	if len(out) == 0 {
		return []domain.Resource{}
	}
	return out
}

func cloneFilteredResource(resource domain.Resource) domain.Resource {
	resource.DependsOn = cloneDependencyIDs(resource.DependsOn)
	resource.DesiredState = cloneDesiredState(resource.DesiredState)
	return resource
}

func cloneDependencyIDs(ids []string) []string {
	if ids == nil {
		return nil
	}
	return append([]string(nil), ids...)
}

func cloneDesiredState(state []byte) []byte {
	if state == nil {
		return nil
	}
	return append([]byte(nil), state...)
}
