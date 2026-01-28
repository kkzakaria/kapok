package deploy

import (
	"testing"

	"github.com/kapok/kapok/internal/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDetector returns a fixed cloud provider.
type MockDetector struct {
	Provider k8s.CloudProvider
}

func (m *MockDetector) Detect() k8s.CloudProvider {
	return m.Provider
}

// MockRunner records commands without executing them.
type MockRunner struct {
	Calls  []MockCall
	Output string
	Err    error
}

type MockCall struct {
	Name string
	Args []string
}

func (m *MockRunner) Run(name string, args ...string) (string, error) {
	m.Calls = append(m.Calls, MockCall{Name: name, Args: args})
	return m.Output, m.Err
}

func TestDeployer_DryRun(t *testing.T) {
	runner := &MockRunner{}
	deployer := &Deployer{
		Detector:  &MockDetector{Provider: k8s.CloudAWS},
		Generator: &k8s.HelmChartGenerator{},
		Runner:    runner,
	}

	dir := t.TempDir()
	err := deployer.Run(Options{
		Namespace: "test",
		Domain:    "test.local",
		ImageTag:  "v1",
		OutputDir: dir,
		DryRun:    true,
		HPA:       true,
	})
	require.NoError(t, err)

	// Helm should not have been called
	assert.Empty(t, runner.Calls)
}

func TestDeployer_FullDeploy(t *testing.T) {
	runner := &MockRunner{Output: "release deployed"}
	deployer := &Deployer{
		Detector:  &MockDetector{Provider: k8s.CloudGCP},
		Generator: &k8s.HelmChartGenerator{},
		Runner:    runner,
	}

	dir := t.TempDir()
	err := deployer.Run(Options{
		Namespace: "prod",
		Domain:    "app.com",
		ImageTag:  "v2",
		OutputDir: dir,
		DryRun:    false,
	})
	require.NoError(t, err)

	// Helm should have been called once
	require.Len(t, runner.Calls, 1)
	call := runner.Calls[0]
	assert.Equal(t, "helm", call.Name)
	assert.Contains(t, call.Args, "upgrade")
	assert.Contains(t, call.Args, "--install")
	assert.Contains(t, call.Args, "--namespace")
	assert.Contains(t, call.Args, "prod")
}

func TestDeployer_ExplicitCloud(t *testing.T) {
	runner := &MockRunner{Output: "ok"}
	deployer := &Deployer{
		Detector:  &MockDetector{Provider: k8s.CloudGeneric},
		Generator: &k8s.HelmChartGenerator{},
		Runner:    runner,
	}

	dir := t.TempDir()
	err := deployer.Run(Options{
		Cloud:     "azure",
		Namespace: "kapok",
		Domain:    "kapok.local",
		ImageTag:  "latest",
		OutputDir: dir,
		DryRun:    true,
	})
	require.NoError(t, err)
}

func TestDeployer_HelmFailure(t *testing.T) {
	runner := &MockRunner{Err: assert.AnError}
	deployer := &Deployer{
		Detector:  &MockDetector{Provider: k8s.CloudAWS},
		Generator: &k8s.HelmChartGenerator{},
		Runner:    runner,
	}

	dir := t.TempDir()
	err := deployer.Run(Options{
		Namespace: "test",
		Domain:    "test.local",
		ImageTag:  "v1",
		OutputDir: dir,
		DryRun:    false,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "helm deploy failed")
}
