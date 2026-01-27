package rbac

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/rs/zerolog"
)

// Enforcer wraps Casbin enforcer with custom methods
type Enforcer struct {
	*casbin.Enforcer
	logger zerolog.Logger
}

// Config holds RBAC configuration
type Config struct {
	ModelPath      string
	DatabaseDSN    string
	TableName      string
}

// NewEnforcer creates a new RBAC enforcer with PostgreSQL adapter
func NewEnforcer(config Config, logger zerolog.Logger) (*Enforcer, error) {
	logger.Info().
		Str("model_path", config.ModelPath).
		Msg("initializing RBAC enforcer")

	// Create GORM adapter for PostgreSQL
	adapter, err := gormadapter.NewAdapter("postgres", config.DatabaseDSN, config.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to create GORM adapter: %w", err)
	}

	// Create Casbin enforcer
	enforcer, err := casbin.NewEnforcer(config.ModelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Load policies from database
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load policies: %w", err)
	}

	logger.Info().Msg("RBAC enforcer initialized successfully")

	return &Enforcer{
		Enforcer: enforcer,
		logger:   logger,
	}, nil
}

// Enforce checks if a subject has permission for an object, action, and tenant
func (e *Enforcer) Enforce(subject, object, action, tenantID string) (bool, error) {
	allowed, err := e.Enforcer.Enforce(subject, object, action, tenantID)
	if err != nil {
		e.logger.Error().
			Err(err).
			Str("subject", subject).
			Str("object", object).
			Str("action", action).
			Str("tenant", tenantID).
			Msg("enforcement check failed")
		return false, fmt.Errorf("enforcement check failed: %w", err)
	}

	e.logger.Debug().
		Str("subject", subject).
		Str("object", object).
		Str("action", action).
		Str("tenant", tenantID).
		Bool("allowed", allowed).
		Msg("enforcement check")

	return allowed, nil
}

// AddPolicy adds a policy rule
func (e *Enforcer) AddPolicy(subject, object, action, tenantID string) error {
	added, err := e.Enforcer.AddPolicy(subject, object, action, tenantID)
	if err != nil {
		return fmt.Errorf("failed to add policy: %w", err)
	}

	if !added {
		e.logger.Warn().
			Str("subject", subject).
			Str("object", object).
			Str("action", action).
			Str("tenant", tenantID).
			Msg("policy already exists")
		return nil
	}

	e.logger.Info().
		Str("subject", subject).
		Str("object", object).
		Str("action", action).
		Str("tenant", tenantID).
		Msg("policy added")

	return nil
}

// RemovePolicy removes a policy rule
func (e *Enforcer) RemovePolicy(subject, object, action, tenantID string) error {
	removed, err := e.Enforcer.RemovePolicy(subject, object, action, tenantID)
	if err != nil {
		return fmt.Errorf("failed to remove policy: %w", err)
	}

	if !removed {
		e.logger.Warn().
			Str("subject", subject).
			Str("object", object).
			Str("action", action).
			Str("tenant", tenantID).
			Msg("policy does not exist")
		return nil
	}

	e.logger.Info().
		Str("subject", subject).
		Str("object", object).
		Str("action", action).
		Str("tenant", tenantID).
		Msg("policy removed")

	return nil
}

// AddRoleForUser adds a role to a user
func (e *Enforcer) AddRoleForUser(userID, role string) error {
	added, err := e.Enforcer.AddRoleForUser(userID, role)
	if err != nil {
		return fmt.Errorf("failed to add role for user: %w", err)
	}

	if !added {
		e.logger.Warn().
			Str("user", userID).
			Str("role", role).
			Msg("role already assigned to user")
		return nil
	}

	e.logger.Info().
		Str("user", userID).
		Str("role", role).
		Msg("role added for user")

	return nil
}

// RemoveRoleForUser removes a role from a user
func (e *Enforcer) RemoveRoleForUser(userID, role string) error {
	removed, err := e.Enforcer.DeleteRoleForUser(userID, role)
	if err != nil {
		return fmt.Errorf("failed to remove role for user: %w", err)
	}

	if !removed {
		e.logger.Warn().
			Str("user", userID).
			Str("role", role).
			Msg("role not assigned to user")
		return nil
	}

	e.logger.Info().
		Str("user", userID).
		Str("role", role).
		Msg("role removed from user")

	return nil
}

// GetRolesForUser gets all roles for a user
func (e *Enforcer) GetRolesForUser(userID string) ([]string, error) {
	roles, err := e.Enforcer.GetRolesForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for user: %w", err)
	}
	return roles, nil
}

// GetUsersForRole gets all users with a specific role
func (e *Enforcer) GetUsersForRole(role string) ([]string, error) {
	users, err := e.Enforcer.GetUsersForRole(role)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for role: %w", err)
	}
	return users, nil
}

// GetPermissionsForUser gets all permissions for a user
func (e *Enforcer) GetPermissionsForUser(userID string) ([][]string, error) {
	permissions, err := e.Enforcer.GetPermissionsForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions for user: %w", err)
	}
	return permissions, nil
}

// SavePolicy saves all policies back to the database
func (e *Enforcer) SavePolicy() error {
	if err := e.Enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("failed to save policies: %w", err)
	}

	e.logger.Info().Msg("policies saved to database")
	return nil
}
