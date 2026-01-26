package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/kapok/kapok/pkg/logger"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start local development environment",
	Long: `Start local development environment with PostgreSQL database.

Requires Docker to be installed and running.
Creates a PostgreSQL container accessible on localhost:5432.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Initialize logger
		logger.Init(logger.Config{
			Level:  "info",
			Format: "console",
		})

		log := logger.Log

		log.Info().Msg("🚀 Starting Kapok development environment")

		// Check if Docker is installed
		if err := checkDockerInstalled(); err != nil {
			return err
		}

		log.Info().Msg("✓ Docker is installed and running")

		// Setup dockertest
		pool, err := dockertest.NewPool("")
		if err != nil {
			return fmt.Errorf("failed to connect to Docker: %w\nIs Docker running?", err)
		}

		// Ping Docker to verify connection
		if err := pool.Client.Ping(); err != nil {
			return fmt.Errorf("failed to ping Docker: %w\nIs Docker running?", err)
		}

		// PostgreSQL configuration
		pgPassword := getEnvOrDefault("KAPOK_DATABASE_PASSWORD", "kapok_dev_password")
		pgUser := getEnvOrDefault("KAPOK_DATABASE_USER", "kapok")
		pgDatabase := getEnvOrDefault("KAPOK_DATABASE_DATABASE", "kapok")

		log.Info().
			Str("user", pgUser).
			Str("database", pgDatabase).
			Msg("Starting PostgreSQL container")

		// Try to find existing container first
		resource, existed, err := findOrCreatePostgres(pool, pgUser, pgPassword, pgDatabase)
		if err != nil {
			return fmt.Errorf("failed to start PostgreSQL: %w", err)
		}

		if existed {
			log.Info().Msg("✓ Using existing PostgreSQL container")
		} else {
			log.Info().Msg("✓ Created new PostgreSQL container")
		}

		// Get connection info
		host := "localhost"
		port := resource.GetPort("5432/tcp")

		// Build connection string (with masked password for display)
		connStr := fmt.Sprintf("postgresql://%s:***@%s:%s/%s?sslmode=disable",
			pgUser, host, port, pgDatabase)
		log.Info().
			Str("connection", connStr).
			Msg("PostgreSQL is ready")

		// Wait for PostgreSQL to be ready
		realConnStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			pgUser, pgPassword, host, port, pgDatabase)

		if err := waitForPostgres(ctx, pool, realConnStr); err != nil {
			return fmt.Errorf("PostgreSQL failed to become ready: %w", err)
		}

		log.Info().Msg("✓ PostgreSQL health check passed")

		// Display next steps
		log.Info().Msg("\n✨ Development environment ready!")
		log.Info().Msg("\n📝 Connection details:")
		log.Info().Msgf("  Host: %s", host)
		log.Info().Msgf("  Port: %s", port)
		log.Info().Msgf("  User: %s", pgUser)
		log.Info().Msgf("  Password: %s", pgPassword)
		log.Info().Msgf("  Database: %s", pgDatabase)
		log.Info().Msg("\n💡 Press Ctrl+C to stop")

		// Keep running until interrupted
		select {}
	},
}

func init() {
	rootCmd.AddCommand(devCmd)
}

// checkDockerInstalled checks if Docker is installed and running
func checkDockerInstalled() error {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(`Docker is not installed or not running.

Please install Docker:
  • macOS: https://docs.docker.com/desktop/install/mac-install/
  • Linux: https://docs.docker.com/engine/install/
  • Windows: https://docs.docker.com/desktop/install/windows-install/

Then start Docker and try again.`)
	}
	return nil
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// findOrCreatePostgres finds existing PostgreSQL container or creates new one
func findOrCreatePostgres(pool *dockertest.Pool, user, password, database string) (*dockertest.Resource, bool, error) {
	containerName := "kapok-postgres-dev"

	// Try to find existing container by name
	containers, err := pool.Client.ListContainers(docker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"name": {containerName},
		},
	})

	if err != nil {
		return nil, false, fmt.Errorf("failed to list containers: %w", err)
	}

	// If container exists, start it if needed and return
	if len(containers) > 0 {
		container := containers[0]

		// If not running, start it
		if container.State != "running" {
			if err := pool.Client.StartContainer(container.ID, nil); err != nil {
				return nil, false, fmt.Errorf("failed to start existing container: %w", err)
			}
		}

		// Find the resource by inspecting the container
		resource, err := pool.Client.InspectContainer(container.ID)
		if err != nil {
			return nil, false, fmt.Errorf("failed to inspect container: %w", err)
		}

		// Wrap in dockertest Resource
		dockerResource := &dockertest.Resource{
			Container: resource,
		}

		return dockerResource, true, nil
	}

	// Create new container with fixed port binding
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:       containerName,
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + database,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432/tcp": {{HostIP: "127.0.0.1", HostPort: "5432"}},
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = false
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		return nil, false, fmt.Errorf("failed to create container: %w", err)
	}

	return resource, false, nil
}

// waitForPostgres waits for PostgreSQL to be ready
func waitForPostgres(ctx context.Context, pool *dockertest.Pool, connStr string) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	return pool.Retry(func() error {
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return err
		}
		defer db.Close()

		return db.PingContext(ctx)
	})
}
