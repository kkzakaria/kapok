package tenant

import (
	"context"
	"fmt"
	"sync"

	"github.com/kapok/kapok/internal/database"
)

// IsolationStrategy defines the interface for tenant isolation approaches
type IsolationStrategy interface {
	Provision(ctx context.Context, tenant *Tenant) error
	Deprovision(ctx context.Context, tenant *Tenant) error
	GetConnection(ctx context.Context, tenant *Tenant) (*database.DB, error)
	Type() string
}

// StrategyRegistry holds registered isolation strategies
type StrategyRegistry struct {
	mu         sync.RWMutex
	strategies map[string]IsolationStrategy
}

// NewStrategyRegistry creates a new strategy registry
func NewStrategyRegistry() *StrategyRegistry {
	return &StrategyRegistry{
		strategies: make(map[string]IsolationStrategy),
	}
}

// Register adds an isolation strategy
func (r *StrategyRegistry) Register(strategy IsolationStrategy) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.strategies[strategy.Type()] = strategy
}

// Get returns the strategy for the given isolation type
func (r *StrategyRegistry) Get(isolationType string) (IsolationStrategy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.strategies[isolationType]
	if !ok {
		return nil, fmt.Errorf("unknown isolation strategy: %s", isolationType)
	}
	return s, nil
}
