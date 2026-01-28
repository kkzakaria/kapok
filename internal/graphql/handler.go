package graphql

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/graphql-go/graphql"
	gqlhandler "github.com/graphql-go/handler"
	"github.com/kapok/kapok/internal/database"
	"github.com/kapok/kapok/internal/tenant"
	"github.com/rs/zerolog"
)

const (
	// SchemaCacheTTL is the time-to-live for cached schemas
	SchemaCacheTTL = 5 * time.Minute
)

// validIdentifier matches valid PostgreSQL identifiers
var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// cachedSchema holds a schema with its expiration time
type cachedSchema struct {
	schema    *graphql.Schema
	expiresAt time.Time
}

// Handler serves GraphQL requests with dynamic schema generation
type Handler struct {
	introspector *Introspector
	generator    *SchemaGenerator
	logger       zerolog.Logger

	// In-memory cache for schemas with TTL
	// Key: schemaName, Value: *cachedSchema
	schemaCache sync.Map
}

// NewHandler creates a new GraphQL handler
func NewHandler(db *database.DB, logger zerolog.Logger) *Handler {
	resolver := NewResolver(db)
	return &Handler{
		introspector: NewIntrospector(db),
		generator:    NewSchemaGenerator(resolver),
		logger:       logger,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Get Tenant Context using the tenant package
	t, err := tenant.GetTenant(ctx)
	if err != nil {
		h.logger.Warn().Err(err).Msg("tenant not found in context")
		http.Error(w, "unauthorized: tenant context required", http.StatusUnauthorized)
		return
	}

	schemaName := t.SchemaName
	if schemaName == "" {
		h.logger.Error().Str("tenant_id", t.ID).Msg("tenant has no schema name")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Validate schema name to prevent injection
	if !validIdentifier.MatchString(schemaName) {
		h.logger.Error().Str("schema_name", schemaName).Msg("invalid schema name format")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug().
		Str("tenant_id", t.ID).
		Str("schema_name", schemaName).
		Msg("handling graphql request")

	// 2. Get Schema (Cache or Generate)
	schema, err := h.getSchema(ctx, schemaName)
	if err != nil {
		h.logger.Error().Err(err).Str("schema_name", schemaName).Msg("failed to get schema")
		http.Error(w, "failed to load schema", http.StatusInternalServerError)
		return
	}

	// 3. Serve via standard handler
	gh := gqlhandler.New(&gqlhandler.Config{
		Schema:   schema,
		Pretty:   true,
		GraphiQL: false,
	})

	gh.ServeHTTP(w, r)
}

func (h *Handler) getSchema(ctx context.Context, schemaName string) (*graphql.Schema, error) {
	// Check cache with TTL
	if val, ok := h.schemaCache.Load(schemaName); ok {
		cached := val.(*cachedSchema)
		if time.Now().Before(cached.expiresAt) {
			return cached.schema, nil
		}
		// Cache expired, remove it
		h.schemaCache.Delete(schemaName)
		h.logger.Debug().Str("schema_name", schemaName).Msg("schema cache expired")
	}

	start := time.Now()

	// Introspect
	metadata, err := h.introspector.Inspect(ctx, schemaName)
	if err != nil {
		return nil, fmt.Errorf("introspection failed: %w", err)
	}

	// Generate
	schema, err := h.generator.Generate(schemaName, metadata)
	if err != nil {
		return nil, fmt.Errorf("schema generation failed: %w", err)
	}

	// Cache with TTL
	h.schemaCache.Store(schemaName, &cachedSchema{
		schema:    schema,
		expiresAt: time.Now().Add(SchemaCacheTTL),
	})

	h.logger.Info().
		Str("schema_name", schemaName).
		Int("tables", len(metadata.Tables)).
		Dur("duration", time.Since(start)).
		Dur("cache_ttl", SchemaCacheTTL).
		Msg("schema generated and cached")

	return schema, nil
}

// InvalidateCache clears the cache for a tenant (e.g. on DDL webhook)
func (h *Handler) InvalidateCache(schemaName string) {
	h.schemaCache.Delete(schemaName)
}
