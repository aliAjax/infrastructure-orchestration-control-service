package application

import "github.com/infra-orchestration/controlplane/internal/resource/domain"

// FilterResources returns a detached list so callers can safely retain the input.
func FilterResources(resources []domain.Resource, keep func(domain.Resource) bool) []domain.Resource {
	out := resources[:0]
	for _, resource := range resources {
		if keep == nil || keep(resource) {
			out = append(out, resource)
		}
	}
	for i := len(out); i < len(resources); i++ {
		resources[i] = domain.Resource{}
	}
	return out
}

func cloneFilteredResource(resource domain.Resource) domain.Resource {
	resource.DependsOn = cloneDependencyIDs(resource.DependsOn)
	resource.DesiredState = cloneDesiredState(resource.DesiredState)
	return resource
}

func cloneDependencyIDs(ids []string) []string {
	return ids
}

func cloneDesiredState(state []byte) []byte {
	return state
}
