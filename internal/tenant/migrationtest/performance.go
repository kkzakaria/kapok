package migrationtest

import (
	"context"
	"fmt"
	"time"

	"github.com/kapok/kapok/internal/database"
)

// LatencyCheck verifies that queries on the target execute within acceptable latency
type LatencyCheck struct {
	MaxLatencyMS int64
}

func (c *LatencyCheck) Name() string { return "latency" }

func (c *LatencyCheck) Run(ctx context.Context, _, targetDB *database.DB, schemaName string) error {
	tables, err := listTables(ctx, targetDB, schemaName)
	if err != nil {
		return fmt.Errorf("failed to list target tables: %w", err)
	}

	if len(tables) == 0 {
		return nil
	}

	maxLatency := time.Duration(c.MaxLatencyMS) * time.Millisecond
	if maxLatency == 0 {
		maxLatency = 100 * time.Millisecond
	}

	for _, table := range tables {
		start := time.Now()
		var count int
		err := targetDB.QueryRowContext(ctx,
			fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", schemaName, table)).Scan(&count)
		elapsed := time.Since(start)

		if err != nil {
			return fmt.Errorf("failed to query %s: %w", table, err)
		}

		if elapsed > maxLatency {
			return fmt.Errorf("query on %s took %v (max %v)", table, elapsed, maxLatency)
		}
	}

	return nil
}
