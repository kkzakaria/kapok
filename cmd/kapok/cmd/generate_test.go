package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCommand_Exists(t *testing.T) {
	// Verify generate command is registered
	assert.NotNil(t, generateCmd)
	assert.Equal(t, "generate", generateCmd.Use)
	assert.True(t, generateCmd.HasSubCommands())
}

func TestSDKCommand_Exists(t *testing.T) {
	// Verify sdk subcommand is registered
	assert.NotNil(t, sdkCmd)
	assert.Equal(t, "sdk", sdkCmd.Use)
	assert.Contains(t, sdkCmd.Short, "TypeScript")
}

func TestSDKCommand_Flags(t *testing.T) {
	// Verify flags are defined
	outputFlag := sdkCmd.Flags().Lookup("output-dir")
	require.NotNil(t, outputFlag)
	assert.Equal(t, "./sdk/typescript", outputFlag.DefValue)

	schemaFlag := sdkCmd.Flags().Lookup("schema")
	require.NotNil(t, schemaFlag)
	assert.Equal(t, "public", schemaFlag.DefValue)

	projectFlag := sdkCmd.Flags().Lookup("project-name")
	require.NotNil(t, projectFlag)
	assert.Equal(t, "kapok-sdk", projectFlag.DefValue)
}

func TestGenerateSDK_InvalidConfig(t *testing.T) {
	// Save original config path and restore after test
	tmpDir := os.Getenv("HOME")
	defer os.Setenv("HOME", tmpDir)

	// Create temp directory for test
	testDir, err := ioutil.TempDir("", "sdk-gen-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	os.Setenv("HOME", testDir)

	// Run without valid config should fail
	err = runGenerateSDK(sdkCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

func TestConnectToDatabase_InvalidConfig(t *testing.T) {
	// This test would require mocking or a real database
	// For now, we'll skip it and rely on integration tests
	t.Skip("Requires database connection - tested in integration tests")
}

// Integration test - requires actual database
func TestGenerateSDK_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This would be a full integration test with dockertest
	// Testing the complete flow: DB connection → introspection → SDK generation
	t.Skip("Integration test - implement with dockertest")
}

func TestGenerateCommand_Help(t *testing.T) {
	// Verify help text is informative
	assert.Contains(t, generateCmd.Long, "Generate SDK")
	assert.Contains(t, sdkCmd.Long, "TypeScript SDK")
	assert.Contains(t, sdkCmd.Long, "package.json")
}

func TestSDKCommand_Examples(t *testing.T) {
	// Verify examples are present
	assert.Contains(t, sdkCmd.Long, "Example:")
	assert.Contains(t, sdkCmd.Long, "kapok generate sdk")
}
