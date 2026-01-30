package migrationtest

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewTestSuite(t *testing.T) {
	logger := zerolog.Nop()
	ts := NewTestSuite(logger)
	assert.NotNil(t, ts)
	assert.Empty(t, ts.checks)
}

func TestAddCheck(t *testing.T) {
	logger := zerolog.Nop()
	ts := NewTestSuite(logger)
	ts.AddCheck(&RowCountCheck{})
	ts.AddCheck(&SchemaCompareCheck{})
	assert.Len(t, ts.checks, 2)
}

func TestDefaultSuite(t *testing.T) {
	logger := zerolog.Nop()
	ts := DefaultSuite(logger)
	assert.Len(t, ts.checks, 3) // RowCount, SchemaCompare, Latency
}

func TestFormatResults(t *testing.T) {
	sr := &SuiteResult{
		Passed: true,
		Results: []Result{
			{CheckName: "a", Passed: true},
			{CheckName: "b", Passed: true},
		},
	}
	assert.Contains(t, FormatResults(sr), "2/2")

	sr.Results = append(sr.Results, Result{CheckName: "c", Passed: false, Error: "fail"})
	sr.Passed = false
	assert.Contains(t, FormatResults(sr), "2/3")
}
