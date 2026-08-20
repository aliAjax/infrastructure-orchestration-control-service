package infrastructure

import (
	"sync"

	graphdomain "github.com/infra-orchestration/controlplane/internal/graph/domain"
)

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]*graphdomain.Graph
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{items: make(map[string]*graphdomain.Graph)}
}

func (c *MemoryCache) Set(key string, graph *graphdomain.Graph) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = graph
}

func (c *MemoryCache) Get(key string) (*graphdomain.Graph, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	g, ok := c.items[key]
	return g, ok
}
