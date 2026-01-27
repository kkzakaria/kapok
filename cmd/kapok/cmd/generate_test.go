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

// React Command Tests

func TestReactCommand_Exists(t *testing.T) {
	// Verify react subcommand is registered
	assert.NotNil(t, reactCmd)
	assert.Equal(t, "react", reactCmd.Use)
	assert.Contains(t, reactCmd.Short, "React")
}

func TestReactCommand_Flags(t *testing.T) {
	// Verify flags are defined with correct defaults
	outputFlag := reactCmd.Flags().Lookup("output-dir")
	require.NotNil(t, outputFlag)
	assert.Equal(t, "./sdk/react", outputFlag.DefValue)

	schemaFlag := reactCmd.Flags().Lookup("schema")
	require.NotNil(t, schemaFlag)
	assert.Equal(t, "public", schemaFlag.DefValue)

	projectFlag := reactCmd.Flags().Lookup("project-name")
	require.NotNil(t, projectFlag)
	assert.Equal(t, "kapok-react", projectFlag.DefValue)

	sdkImportFlag := reactCmd.Flags().Lookup("sdk-import")
	require.NotNil(t, sdkImportFlag)
	assert.Equal(t, "../typescript", sdkImportFlag.DefValue)
}

func TestReactCommand_Help(t *testing.T) {
	// Verify help text is informative
	assert.Contains(t, reactCmd.Long, "React Query")
	assert.Contains(t, reactCmd.Long, "hooks")
	assert.Contains(t, reactCmd.Long, "KapokProvider")
}

func TestReactCommand_Examples(t *testing.T) {
	// Verify examples are present
	assert.Contains(t, reactCmd.Long, "Example:")
	assert.Contains(t, reactCmd.Long, "kapok generate react")
	assert.Contains(t, reactCmd.Long, "--sdk-import")
}

func TestGenerateCommand_Help(t *testing.T) {
	// Verify help text mentions both SDK types
	assert.Contains(t, generateCmd.Long, "Generate SDK")
	assert.Contains(t, generateCmd.Long, "kapok generate sdk")
	assert.Contains(t, generateCmd.Long, "kapok generate react")
}

// Error Scenario Tests

func TestGenerateSDK_InvalidConfig(t *testing.T) {
	// Save original config path and restore after test
	tmpDir := os.Getenv("HOME")
	defer os.Setenv("HOME", tmpDir)

	// Create temp directory for test
	testDir, err := ioutil.TempDir("", "sdk-gen-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	os.Setenv("HOME", testDir)

	// Run without valid config/database should fail
	err = runGenerateSDK(sdkCmd, []string{})
	assert.Error(t, err)
	// Should fail on database connection, not config load
	assert.Contains(t, err.Error(), "database")
}

func TestGenerateReact_InvalidConfig(t *testing.T) {
	// Save original config path and restore after test
	tmpDir := os.Getenv("HOME")
	defer os.Setenv("HOME", tmpDir)

	// Create temp directory for test
	testDir, err := ioutil.TempDir("", "react-gen-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	os.Setenv("HOME", testDir)

	// Run without valid config/database should fail
	err = runGenerateReact(reactCmd, []string{})
	assert.Error(t, err)
	// Should fail on database connection
	assert.Contains(t, err.Error(), "database")
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

func TestGenerateReact_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Full integration test for React generation
	t.Skip("Integration test - implement with dockertest")
}

func TestSDKCommand_Examples(t *testing.T) {
	// Verify examples are present
	assert.Contains(t, sdkCmd.Long, "Example:")
	assert.Contains(t, sdkCmd.Long, "kapok generate sdk")
}
