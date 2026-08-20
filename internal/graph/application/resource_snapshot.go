package application

import "github.com/infra-orchestration/controlplane/internal/resource/domain"

// cloneResourceSnapshot returns a deep copy of resources so concurrent
// execution layers never share nested state (DependsOn, DesiredState) with
// the caller's input. A plain slice copy is not enough: the nested slices
// would still alias the source and let one worker's writes leak into another.
func cloneResourceSnapshot(resources []domain.Resource) []domain.Resource {
	if resources == nil {
		return nil
	}
	cloned := make([]domain.Resource, len(resources))
	for i, r := range resources {
		cloned[i] = r
		if r.DependsOn != nil {
			cloned[i].DependsOn = append([]string(nil), r.DependsOn...)
		}
		if r.DesiredState != nil {
			cloned[i].DesiredState = append([]byte(nil), r.DesiredState...)
		}
	}
	return cloned
}
