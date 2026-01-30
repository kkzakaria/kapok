package tenant

import (
	"sync"

	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
)

// PoolManager manages per-tenant database connection pools
type PoolManager struct {
	pools  map[string]*database.DB
	mu     sync.RWMutex
	logger zerolog.Logger
}

// NewPoolManager creates a new PoolManager
func NewPoolManager(logger zerolog.Logger) *PoolManager {
	return &PoolManager{
		pools:  make(map[string]*database.DB),
		logger: logger,
	}
}

// Get returns the connection pool for a tenant
func (pm *PoolManager) Get(tenantID string) (*database.DB, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	db, ok := pm.pools[tenantID]
	return db, ok
}

// Set registers a connection pool for a tenant
func (pm *PoolManager) Set(tenantID string, db *database.DB) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.pools[tenantID] = db
	pm.logger.Info().Str("tenant_id", tenantID).Msg("registered tenant connection pool")
}

// Remove closes and removes a tenant's connection pool
func (pm *PoolManager) Remove(tenantID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if db, ok := pm.pools[tenantID]; ok {
		if err := db.Close(); err != nil {
			pm.logger.Error().Err(err).Str("tenant_id", tenantID).Msg("failed to close tenant pool")
		}
		delete(pm.pools, tenantID)
		pm.logger.Info().Str("tenant_id", tenantID).Msg("removed tenant connection pool")
	}
}

// CloseAll closes all connection pools
func (pm *PoolManager) CloseAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for id, db := range pm.pools {
		if err := db.Close(); err != nil {
			pm.logger.Error().Err(err).Str("tenant_id", id).Msg("failed to close tenant pool")
		}
	}
	pm.pools = make(map[string]*database.DB)
	pm.logger.Info().Msg("all tenant connection pools closed")
}

// Count returns the number of active pools
func (pm *PoolManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.pools)
}
