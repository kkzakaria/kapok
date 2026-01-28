package graphql

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/graphql-go/graphql"
	gqlhandler "github.com/graphql-go/handler"
	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
)

// Handler serves GraphQL requests with dynamic schema generation
type Handler struct {
	introspector *Introspector
	generator    *SchemaGenerator
	logger       zerolog.Logger
	
	// Simple in-memory cache for schemas
	// Key: schemaName (tenant_id), Value: *graphql.Schema
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
	// 1. Get Tenant Context
	// In a real middleware, tenant_id would be in context.
	// For MVP/Verification, if context is missing, we might error or fallback (unlikely).
	// Let's assume the router middleware puts "tenant_id" and "schema_name" in context.
	// But `internal/tenant/router.go` typically handles this.
	// For now, let's extract it or fail.
	
	tenantID, ok := r.Context().Value("tenant_id").(string)
	if !ok || tenantID == "" {
		http.Error(w, "tenant_id context required", http.StatusUnauthorized)
		return
	}
	
	schemaName, ok := r.Context().Value("schema_name").(string)
	if !ok || schemaName == "" {
		// Fallback naming convention: tenant_<id> or just use ID if simple
		// Ideally middleware sets this. 
		http.Error(w, "schema_name context required", http.StatusInternalServerError)
		return
	}

	h.logger.Debug().
		Str("tenant_id", tenantID).
		Str("schema_name", schemaName).
		Msg("handling graphql request")

	// 2. Get Schema (Cache or Generate)
	schema, err := h.getSchema(r.Context(), schemaName)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get schema")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Serve via standard handler
	// We create a new handler instance for this request's schema.
	// In high-load, we might want to cache the *handler* itself too, 
	// but *graphql.Schema represents the heavy lifting.
	gh := gqlhandler.New(&gqlhandler.Config{
		Schema:   schema,
		Pretty:   true,
		GraphiQL: false, // Can be enabled via query param or env
	})

	gh.ServeHTTP(w, r)
}

func (h *Handler) getSchema(ctx context.Context, schemaName string) (*graphql.Schema, error) {
	// Check cache
	if val, ok := h.schemaCache.Load(schemaName); ok {
		return val.(*graphql.Schema), nil
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

	// Cache
	h.schemaCache.Store(schemaName, schema)

	h.logger.Info().
		Str("schema_name", schemaName).
		Int("tables", len(metadata.Tables)).
		Dur("duration", time.Since(start)).
		Msg("schema generated and cached")

	return schema, nil
}

// InvalidateCache clears the cache for a tenant (e.g. on DDL webhook)
func (h *Handler) InvalidateCache(schemaName string) {
	h.schemaCache.Delete(schemaName)
}
