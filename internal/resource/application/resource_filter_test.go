package application

import (
	"context"
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

func TestFilterResourcesReturnsConcreteEmptySlice(t *testing.T) {
	filtered := FilterResources(nil, nil)
	if filtered == nil || len(filtered) != 0 {
		t.Fatalf("empty filter result must be non-nil: %#v", filtered)
	}
}

func TestFilterResourcesCopiesDependencyIDs(t *testing.T) {
	resources := []domain.Resource{{ID: "db", DependsOn: []string{"network"}}}
	filtered := FilterResources(resources, nil)
	filtered[0].DependsOn[0] = "mutated"
	if resources[0].DependsOn[0] != "network" {
		t.Fatalf("dependency IDs share storage: %#v", resources[0].DependsOn)
	}
}

func TestFilterResourcesCopiesDesiredState(t *testing.T) {
	resources := []domain.Resource{{ID: "db", DesiredState: []byte(`{"size":1}`)}}
	filtered := FilterResources(resources, nil)
	filtered[0].DesiredState[0] = '['
	if string(resources[0].DesiredState) != `{"size":1}` {
		t.Fatalf("desired state shares storage: %s", resources[0].DesiredState)
	}
}

type listResourcesRepository struct {
	domain.Repository
	resources []domain.Resource
}

func (r listResourcesRepository) ListResources(context.Context, string) ([]domain.Resource, error) {
	return r.resources, nil
}

func TestListResourcesDetachesRepositoryResult(t *testing.T) {
	stored := []domain.Resource{{ID: "db"}}
	got, err := NewService(listResourcesRepository{resources: stored}).ListResources(context.Background(), "prod")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].EnvironmentID != "prod" {
		t.Fatalf("environment identity was not normalized: %#v", got)
	}
	got[0].ID = "changed"
	if stored[0].ID != "db" {
		t.Fatalf("service returned repository storage directly: %#v", stored)
	}
}
