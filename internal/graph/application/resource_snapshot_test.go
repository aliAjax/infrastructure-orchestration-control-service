package application

import (
	"github.com/infra-orchestration/controlplane/internal/resource/domain"
	"testing"
)

func TestCloneResourceSnapshotDeepCopiesNestedState(t *testing.T) {
	source := []domain.Resource{{ID: "db", DependsOn: []string{"network"}, DesiredState: []byte(`{"size":1}`)}}
	clone := cloneResourceSnapshot(source)
	clone[0].DependsOn[0] = "changed"
	clone[0].DesiredState[0] = '['
	if source[0].DependsOn[0] != "network" || string(source[0].DesiredState) != `{"size":1}` {
		t.Fatalf("snapshot retained nested aliases: %#v", source)
	}
}
