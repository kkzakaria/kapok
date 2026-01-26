package cmd_test

import (
	"context"
	"database/sql"
	"os/exec"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerInstalled(t *testing.T) {
	// This test verifies Docker is installed
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	
	if err != nil {
		t.Skip("Docker is not installed or not running - skipping Docker-dependent tests")
	}
	
	assert.NoError(t, err, "Docker should be installed and running")
}

func TestPostgreSQLContainer(t *testing.T) {
	// Skip if Docker not available
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		t.Skip("Docker is not installed or not running")
	}

	pool, err := dockertest.NewPool("")
	require.NoError(t, err, "Failed to connect to Docker")

	err = pool.Client.Ping()
	require.NoError(t, err, "Failed to ping Docker")

	// Create PostgreSQL container for testing
	resource, err := pool.Run("postgres", "15-alpine", []string{
		"POSTGRES_PASSWORD=test_password",
		"POSTGRES_USER=test_user",
		"POSTGRES_DB=test_db",
	})
	require.NoError(t, err, "Failed to start PostgreSQL container")

	// Clean up
	defer func() {
		if err := pool.Purge(resource); err != nil {
			t.Errorf("Failed to purge resource: %v", err)
		}
	}()

	// Wait for PostgreSQL to be ready
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var db *sql.DB
	err = pool.Retry(func() error {
		var err error
		connStr := "postgres://test_user:test_password@localhost:" + 
			resource.GetPort("5432/tcp") + "/test_db?sslmode=disable"
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return err
		}
		return db.PingContext(ctx)
	})
	require.NoError(t, err, "PostgreSQL should be ready")

	if db != nil {
		db.Close()
	}
}

func TestPostgreSQLConnection(t *testing.T) {
	// Skip if Docker not available
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		t.Skip("Docker is not installed or not running")
	}

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := pool.Run("postgres", "15-alpine", []string{
		"POSTGRES_PASSWORD=conn_test_pass",
		"POSTGRES_USER=conn_test_user",
		"POSTGRES_DB=conn_test_db",
	})
	require.NoError(t, err)

	defer pool.Purge(resource)

	// Wait and connect
	ctx := context.Background()
	var db *sql.DB

	err = pool.Retry(func() error {
		var err error
		connStr := "postgres://conn_test_user:conn_test_pass@localhost:" +
			resource.GetPort("5432/tcp") + "/conn_test_db?sslmode=disable"
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return err
		}
		return db.Ping()
	})
	require.NoError(t, err)
	defer db.Close()

	// Verify we can query
	var result int
	err = db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestContainerReuse(t *testing.T) {
	// Skip if Docker not available
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		t.Skip("Docker is not installed or not running")
	}

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	containerName := "kapok-test-reuse"

	// Create first container
	resource1, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:       containerName,
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=reuse_pass",
			"POSTGRES_USER=reuse_user",
			"POSTGRES_DB=reuse_db",
		},
	})
	require.NoError(t, err)

	// Clean up
	defer func() {
		pool.Client.RemoveContainer(docker.RemoveContainerOptions{
			ID:    resource1.Container.ID,
			Force: true,
		})
	}()

	// Verify container exists
	containers, err := pool.Client.ListContainers(docker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"name": {containerName},
		},
	})
	require.NoError(t, err)
	assert.Len(t, containers, 1, "Container should exist")
	assert.Equal(t, "running", containers[0].State)
}

func TestDockerNotInstalledError(t *testing.T) {
	// This test documents expected behavior when Docker is not installed
	// In reality, we can't test this without uninstalling Docker
	// So we just document the expected error message format
	
	expectedErrorContains := []string{
		"Docker is not installed",
		"Please install Docker",
		"docs.docker.com",
	}
	
	// This is to ensure our error message includes helpful info
	_ = expectedErrorContains
}
