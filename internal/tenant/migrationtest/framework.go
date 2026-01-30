package migrationtest

import (
	"context"
	"fmt"

	"github.com/kapok/kapok/internal/database"
	"github.com/rs/zerolog"
)

// Check defines a single migration verification check
type Check interface {
	Name() string
	Run(ctx context.Context, sourceDB, targetDB *database.DB, schemaName string) error
}

// TestSuite runs a set of checks to verify a migration
type TestSuite struct {
	checks []Check
	logger zerolog.Logger
}

// NewTestSuite creates a new TestSuite
func NewTestSuite(logger zerolog.Logger) *TestSuite {
	return &TestSuite{logger: logger}
}

// AddCheck registers a check
func (ts *TestSuite) AddCheck(c Check) {
	ts.checks = append(ts.checks, c)
}

// Result holds the outcome of a single check
type Result struct {
	CheckName string `json:"check_name"`
	Passed    bool   `json:"passed"`
	Error     string `json:"error,omitempty"`
}

// SuiteResult holds the outcome of a full suite run
type SuiteResult struct {
	Results []Result `json:"results"`
	Passed  bool     `json:"passed"`
}

// Run executes all registered checks
func (ts *TestSuite) Run(ctx context.Context, sourceDB, targetDB *database.DB, schemaName string) *SuiteResult {
	sr := &SuiteResult{Passed: true}

	for _, c := range ts.checks {
		ts.logger.Info().Str("check", c.Name()).Msg("running migration check")
		err := c.Run(ctx, sourceDB, targetDB, schemaName)
		r := Result{CheckName: c.Name(), Passed: err == nil}
		if err != nil {
			r.Error = err.Error()
			sr.Passed = false
			ts.logger.Error().Err(err).Str("check", c.Name()).Msg("check failed")
		} else {
			ts.logger.Info().Str("check", c.Name()).Msg("check passed")
		}
		sr.Results = append(sr.Results, r)
	}

	return sr
}

// DefaultSuite creates a test suite with all standard checks
func DefaultSuite(logger zerolog.Logger) *TestSuite {
	ts := NewTestSuite(logger)
	ts.AddCheck(&RowCountCheck{})
	ts.AddCheck(&SchemaCompareCheck{})
	ts.AddCheck(&LatencyCheck{MaxLatencyMS: 100})
	return ts
}

// FormatResults returns a human-readable summary
func FormatResults(sr *SuiteResult) string {
	passed := 0
	for _, r := range sr.Results {
		if r.Passed {
			passed++
		}
	}
	return fmt.Sprintf("%d/%d checks passed (overall: %v)", passed, len(sr.Results), sr.Passed)
}
