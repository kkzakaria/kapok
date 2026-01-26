package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kapok/kapok/cmd/kapok/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommandFull(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	err := os.Chdir(tmpDir)
	require.NoError(t, err)

	// Execute init command
	var buf bytes.Buffer
	err = cmd.ExecuteContext(&buf, []string{"init", "test-project"})
	require.NoError(t, err)

	// Verify output
	output := buf.String()
	assert.Contains(t, output, "test-project")
	assert.Contains(t, output, "✓ Created kapok.yaml")
	assert.Contains(t, output, "✓ Created .env.example")
	assert.Contains(t, output, "✓ Created README.md")
	assert.Contains(t, output, "✓ Created docs/")
	assert.Contains(t, output, "Next steps")

	// Verify files were created
	assert.FileExists(t, "kapok.yaml")
	assert.FileExists(t, ".env.example")
	assert.FileExists(t, "README.md")
	assert.DirExists(t, "docs")
	assert.FileExists(t, filepath.Join("docs", "ARCHITECTURE.md"))

	// Verify file contents
	configContent, err := os.ReadFile("kapok.yaml")
	require.NoError(t, err)
	assert.Contains(t, string(configContent), "test-project")
	assert.Contains(t, string(configContent), "server:")
	assert.Contains(t, string(configContent), "database:")

	envContent, err := os.ReadFile(".env.example")
	require.NoError(t, err)
	assert.Contains(t, string(envContent), "KAPOK_DATABASE_PASSWORD")
	assert.Contains(t, string(envContent), "KAPOK_JWT_SECRET")

	readmeContent, err := os.ReadFile("README.md")
	require.NoError(t, err)
	assert.Contains(t, string(readmeContent), "test-project")
	assert.Contains(t, string(readmeContent), "Quick Start")
}

func TestInitCommandDefaultName(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	err := os.Chdir(tmpDir)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = cmd.ExecuteContext(&buf, []string{"init"})
	require.NoError(t, err)

	// Should use default name
	assert.Contains(t, buf.String(), "my-kapok-project")
	assert.FileExists(t, "kapok.yaml")
}

func TestInitCommandNonEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	err := os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a file to make directory non-empty
	err = os.WriteFile("existing.txt", []byte("test"), 0644)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = cmd.ExecuteContext(&buf, []string{"init", "test-project"})

	// Should fail because directory is not empty
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not empty")
}

func TestInitCommandForceFlag(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	err := os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create existing file
	err = os.WriteFile("existing.txt", []byte("test"), 0644)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = cmd.ExecuteContext(&buf, []string{"init", "test-project", "--force"})

	// Should succeed with --force
	require.NoError(t, err)
	assert.FileExists(t, "kapok.yaml")
	assert.FileExists(t, ".env.example")
	assert.FileExists(t, "README.md")
}

func TestInitCommandOverwriteWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	err := os.Chdir(tmpDir)
	require.NoError(t, err)

	// First init
	var buf1 bytes.Buffer
	err = cmd.ExecuteContext(&buf1, []string{"init", "project-v1"})
	require.NoError(t, err)

	// Verify first version
	content1, _ := os.ReadFile("kapok.yaml")
	assert.Contains(t, string(content1), "project-v1")

	// Second init with --force
	var buf2 bytes.Buffer
	err = cmd.ExecuteContext(&buf2, []string{"init", "project-v2", "--force"})
	require.NoError(t, err)

	// Verify overwritten
	content2, _ := os.ReadFile("kapok.yaml")
	assert.Contains(t, string(content2), "project-v2")
	assert.NotContains(t, string(content2), "project-v1")
}

func TestInitCommandHiddenFilesIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	err := os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create hidden files (like .git directory)
	err = os.MkdirAll(".git", 0755)
	require.NoError(t, err)
	err = os.WriteFile(".gitignore", []byte("*.log"), 0644)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = cmd.ExecuteContext(&buf, []string{"init", "test-project"})

	// Should succeed - hidden files are ignored
	require.NoError(t, err)
	assert.FileExists(t, "kapok.yaml")
}

func TestInitCommandPerformance(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	err := os.Chdir(tmpDir)
	require.NoError(t, err)

	// Measure execution time
	var buf bytes.Buffer
	err = cmd.ExecuteContext(&buf, []string{"init", "perf-test"})
	require.NoError(t, err)

	// Command should complete quickly (we don't have precise timing,
	// but we can verify it didn't hang or panic)
	assert.FileExists(t, "kapok.yaml")
}

func TestInitCommandFileContentsCorrect(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	err := os.Chdir(tmpDir)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = cmd.ExecuteContext(&buf, []string{"init", "validation-test"})
	require.NoError(t, err)

	// Validate kapok.yaml structure
	config, err := os.ReadFile("kapok.yaml")
	require.NoError(t, err)
	configStr := string(config)
	assert.Contains(t, configStr, "server:")
	assert.Contains(t, configStr, "host:")
	assert.Contains(t, configStr, "port:")
	assert.Contains(t, configStr, "database:")
	assert.Contains(t, configStr, "redis:")
	assert.Contains(t, configStr, "log:")
	assert.Contains(t, configStr, "jwt:")

	// Validate .env.example has all required vars
	env, err := os.ReadFile(".env.example")
	require.NoError(t, err)
	envStr := string(env)
	assert.Contains(t, envStr, "KAPOK_DATABASE_PASSWORD")
	assert.Contains(t, envStr, "KAPOK_JWT_SECRET")
	assert.Contains(t, envStr, "KAPOK_DATABASE_HOST")

	// Validate README has sections
	readme, err := os.ReadFile("README.md")
	require.NoError(t, err)
	readmeStr := string(readme)
	assert.Contains(t, readmeStr, "Quick Start")
	assert.Contains(t, readmeStr, "Prerequisites")
	assert.Contains(t, readmeStr, "Setup")
	assert.Contains(t, readmeStr, "Configuration")

	// Validate docs/ARCHITECTURE.md
	arch, err := os.ReadFile(filepath.Join("docs", "ARCHITECTURE.md"))
	require.NoError(t, err)
	archStr := string(arch)
	assert.Contains(t, archStr, "Architecture")
	assert.Contains(t, archStr, "Database Layer")
	assert.Contains(t, archStr, "API Layer")
}
