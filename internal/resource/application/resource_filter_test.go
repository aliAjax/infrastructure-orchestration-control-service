package application

import (
	"testing"

	"github.com/infra-orchestration/controlplane/internal/resource/domain"
)

func TestFilterResourcesDoesNotPolluteOriginalSlice(t *testing.T) {
	resources := []domain.Resource{{ID: "db", Sensitive: true}, {ID: "cache"}, {ID: "queue"}}
	filtered := FilterResources(resources, func(resource domain.Resource) bool { return resource.Sensitive })
	if len(filtered) != 1 || filtered[0].ID != "db" {
		t.Fatalf("unexpected filtered resources: %#v", filtered)
	}
	if len(resources) != 3 || resources[1].ID != "cache" || resources[2].ID != "queue" {
		t.Fatalf("original resources were polluted: %#v", resources)
	}
}
