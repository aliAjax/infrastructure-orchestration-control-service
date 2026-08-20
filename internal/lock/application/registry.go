package application

import (
	"sort"
	"sync"

	"github.com/infra-orchestration/controlplane/internal/resource/domain"
)

type DeclarationRegistry struct {
	mu    sync.RWMutex
	items map[string][]string
}

func NewDeclarationRegistry() *DeclarationRegistry {
	return &DeclarationRegistry{items: make(map[string][]string)}
}

func (r *DeclarationRegistry) Rebuild(resources []domain.Resource) {
	next := make(map[string][]string)
	for _, resource := range resources {
		if resource.LockKey != "" {
			next[resource.LockKey] = append(next[resource.LockKey], resource.ID)
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = next
}

func (r *DeclarationRegistry) LockKeys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]string, 0, len(r.items))
	for key := range r.items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (r *DeclarationRegistry) ResourcesForKey(key string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]string(nil), r.items[key]...)
}
