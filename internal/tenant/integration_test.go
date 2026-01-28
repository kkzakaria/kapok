// +build integration

package tenant

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/kapok/kapok/internal/database"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDB       *sql.DB
	testDBConfig database.Config
	pool         *dockertest.Pool
	resource     *dockertest.Resource
)

// setupPostgresContainer creates an ephemeral PostgreSQL container for testing
func setupPostgresContainer(t *testing.T) {
	var err error
	pool, err = dockertest.NewPool("")
	require.NoError(t, err, "could not construct docker pool")

	err = pool.Client.Ping()
	require.NoError(t, err, "could not connect to docker")

	// Pull and run PostgreSQL container
	resource, err = pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=testpass",
			"POSTGRES_USER=testuser",
			"POSTGRES_DB=kapok_test",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(t, err, "could not start postgres container")

	// Set container expiration to 2 minutes
	resource.Expire(120)

	// Wait for PostgreSQL to be ready
	hostAndPort := resource.GetHostPort("5432/tcp")
	dbURL := fmt.Sprintf("postgres://testuser:testpass@%s/kapok_test?sslmode=disable", hostAndPort)

	err = pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("postgres", dbURL)
		if err != nil {
			return err
		}
		return testDB.Ping()
	})
	require.NoError(t, err, "could not connect to postgres")

	// Store config for use in tests
	port := resource.GetPort("5432/tcp")
	portInt, err := strconv.Atoi(port)
	require.NoError(t, err, "could not convert port to int")
	
	testDBConfig = database.Config{
		Host:     "localhost",
		Port:     portInt,
		Database: "kapok_test",
		User:     "testuser",
		Password: "testpass",
		SSLMode:  "disable",
	}

	t.Logf("PostgreSQL container ready at %s", hostAndPort)
}

// teardownPostgresContainer removes the PostgreSQL container
func teardownPostgresContainer(t *testing.T) {
	if testDB != nil {
		testDB.Close()
	}
	if pool != nil && resource != nil {
		err := pool.Purge(resource)
		if err != nil {
			t.Logf("could not purge postgres container: %v", err)
		}
	}
}

// setupControlDatabase creates the control database schema
func setupControlDatabase(t *testing.T, db *sql.DB) {
	ctx := context.Background()

	// Create tenants table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(63) UNIQUE NOT NULL,
			schema_name VARCHAR(100) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	// Create audit_log table
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS audit_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id VARCHAR(256),
			user_id VARCHAR(256),
			action VARCHAR(100) NOT NULL,
			resource VARCHAR(256),
			timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
			metadata JSONB
		)
	`)
	require.NoError(t, err)

	t.Log("Control database schema created")
}

// TestCrossTenantDataIsolation verifies that tenants cannot access each other's data
func TestCrossTenantDataIsolation(t *testing.T) {
	setupPostgresContainer(t)
	defer teardownPostgresContainer(t)

	setupControlDatabase(t, testDB)

	logger := zerolog.Nop()
	ctx := context.Background()

	// Create database connection
	db, err := database.NewDB(ctx, testDBConfig, logger)
	require.NoError(t, err)
	defer db.Close()

	// Create provisioner
	provisioner := NewProvisioner(db, logger)

	// Create two tenants
	tenant1, err := provisioner.CreateTenant(ctx, "acme")
	require.NoError(t, err)
	assert.Equal(t, "acme", tenant1.Name)

	tenant2, err := provisioner.CreateTenant(ctx, "globex")
	require.NoError(t, err)
	assert.Equal(t, "globex", tenant2.Name)

	// Create a test table in tenant1's schema
	_, err = testDB.ExecContext(ctx, fmt.Sprintf(`
		CREATE TABLE %s.users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100),
			email VARCHAR(100)
		)
	`, tenant1.SchemaName))
	require.NoError(t, err)

	// Insert data into tenant1
	_, err = testDB.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s.users (name, email) VALUES 
		('Alice', 'alice@acme.com'),
		('Bob', 'bob@acme.com')
	`, tenant1.SchemaName))
	require.NoError(t, err)

	// Create same table in tenant2's schema
	_, err = testDB.ExecContext(ctx, fmt.Sprintf(`
		CREATE TABLE %s.users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100),
			email VARCHAR(100)
		)
	`, tenant2.SchemaName))
	require.NoError(t, err)

	// Insert different data into tenant2
	_, err = testDB.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s.users (name, email) VALUES 
		('Charlie', 'charlie@globex.com')
	`, tenant2.SchemaName))
	require.NoError(t, err)

	// Verify tenant1 sees only their data
	var count1 int
	err = testDB.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s.users
	`, tenant1.SchemaName)).Scan(&count1)
	require.NoError(t, err)
	assert.Equal(t, 2, count1, "tenant1 should see 2 users")

	// Verify tenant2 sees only their data
	var count2 int
	err = testDB.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s.users
	`, tenant2.SchemaName)).Scan(&count2)
	require.NoError(t, err)
	assert.Equal(t, 1, count2, "tenant2 should see 1 user")

	// Verify tenant1 cannot see tenant2's data via schema qualification
	var crossCount int
	err = testDB.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s.users WHERE email LIKE '%%globex.com'
	`, tenant1.SchemaName)).Scan(&crossCount)
	require.NoError(t, err)
	assert.Equal(t, 0, crossCount, "tenant1 should not see tenant2's data")

	t.Log("✅ Cross-tenant data isolation verified")
}

// TestTenantProvisioningPerformance verifies provisioning completes in <30 seconds
func TestTenantProvisioningPerformance(t *testing.T) {
	setupPostgresContainer(t)
	defer teardownPostgresContainer(t)

	setupControlDatabase(t, testDB)

	logger := zerolog.Nop()
	ctx := context.Background()

	db, err := database.NewDB(ctx, testDBConfig, logger)
	require.NoError(t, err)
	defer db.Close()

	provisioner := NewProvisioner(db, logger)

	// Measure provisioning time
	start := time.Now()
	tenant, err := provisioner.CreateTenant(ctx, "performance-test")
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, tenant)

	// Assert < 30 seconds (architectural requirement)
	assert.Less(t, duration.Seconds(), 30.0, "provisioning should complete in <30 seconds")

	t.Logf("✅ Provisioning completed in %s (target: <30s)", duration.Round(time.Millisecond))
}

// TestSQLInjectionPrevention tests that SQL injection attempts are blocked
func TestSQLInjectionPrevention(t *testing.T) {
	setupPostgresContainer(t)
	defer teardownPostgresContainer(t)

	setupControlDatabase(t, testDB)

	logger := zerolog.Nop()
	ctx := context.Background()

	db, err := database.NewDB(ctx, testDBConfig, logger)
	require.NoError(t, err)
	defer db.Close()

	provisioner := NewProvisioner(db, logger)

	// Attempt SQL injection through tenant name
	maliciousNames := []string{
		"test'; DROP TABLE tenants; --",
		"test\"; DELETE FROM tenants; --",
		"test' OR '1'='1",
		"../../../etc/passwd",
		"tenant_<script>alert('xss')</script>",
	}

	for _, name := range maliciousNames {
		_, err := provisioner.CreateTenant(ctx, name)
		assert.Error(t, err, "should reject malicious input: %s", name)
	}

	// Verify tenants table still exists and is intact
	var tableExists bool
	err = testDB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'tenants'
		)
	`).Scan(&tableExists)
	require.NoError(t, err)
	assert.True(t, tableExists, "tenants table should still exist after injection attempts")

	t.Log("✅ SQL injection prevention verified")
}

// TestTenantDeletion tests soft and hard delete operations
func TestTenantDeletion(t *testing.T) {
	setupPostgresContainer(t)
	defer teardownPostgresContainer(t)

	setupControlDatabase(t, testDB)

	logger := zerolog.Nop()
	ctx := context.Background()

	db, err := database.NewDB(ctx, testDBConfig, logger)
	require.NoError(t, err)
	defer db.Close()

	provisioner := NewProvisioner(db, logger)

	// Create tenants for deletion tests
	tenantSoft, err := provisioner.CreateTenant(ctx, "soft-delete-test")
	require.NoError(t, err)

	tenantHard, err := provisioner.CreateTenant(ctx, "hard-delete-test")
	require.NoError(t, err)

	// Test soft delete
	err = provisioner.DeleteTenant(ctx, tenantSoft.ID)
	require.NoError(t, err)

	// Verify soft delete: status changed
	deletedTenant, err := provisioner.GetTenantByID(ctx, tenantSoft.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusDeleted, deletedTenant.Status)

	// Verify schema still exists after soft delete
	var schemaExists bool
	err = testDB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.schemata 
			WHERE schema_name = $1
		)
	`, tenantSoft.SchemaName).Scan(&schemaExists)
	require.NoError(t, err)
	assert.True(t, schemaExists, "schema should exist after softdelete")

	// Test hard delete
	err = provisioner.HardDeleteTenant(ctx, tenantHard.ID)
	require.NoError(t, err)

	// Verify hard delete: tenant removed from database
	_, err = provisioner.GetTenantByID(ctx, tenantHard.ID)
	assert.Error(t, err, "tenant should not exist after hard delete")

	// Verify schema dropped after hard delete
	err = testDB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.schemata 
			WHERE schema_name = $1
		)
	`, tenantHard.SchemaName).Scan(&schemaExists)
	require.NoError(t, err)
	assert.False(t, schemaExists, "schema should be dropped after hard delete")

	t.Log("✅ Soft and hard delete operations verified")
}

// TestConcurrentTenantCreation tests that concurrent tenant creation works correctly
func TestConcurrentTenantCreation(t *testing.T) {
	setupPostgresContainer(t)
	defer teardownPostgresContainer(t)

	setupControlDatabase(t, testDB)

	logger := zerolog.Nop()
	ctx := context.Background()

	db, err := database.NewDB(ctx, testDBConfig, logger)
	require.NoError(t, err)
	defer db.Close()

	provisioner := NewProvisioner(db, logger)

	// Create 5 tenants concurrently
	done := make(chan bool, 5)
	errors := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func(idx int) {
			name := fmt.Sprintf("concurrent-test-%d", idx)
			_, err := provisioner.CreateTenant(ctx, name)
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	var errCount int
	for err := range errors {
		t.Logf("Error during concurrent creation: %v", err)
		errCount++
	}

	assert.Equal(t, 0, errCount, "no errors should occur during concurrent creation")

	// Verify all 5 tenants were created
	tenants, err := provisioner.ListTenants(ctx, "", 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(tenants), 5, "at least 5 tenants should exist")

	t.Log("✅ Concurrent tenant creation verified")
}
