package application

import "github.com/infra-orchestration/controlplane/internal/resource/domain"

// FilterResources returns a detached list so callers can safely retain the input.
func FilterResources(resources []domain.Resource, keep func(domain.Resource) bool) []domain.Resource {
	out := make([]domain.Resource, 0, len(resources))
	for _, resource := range resources {
		if keep == nil || keep(resource) {
			out = append(out, resource)
		}
	}
	return out
}
