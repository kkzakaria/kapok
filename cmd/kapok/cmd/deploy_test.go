package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeployCommand_Exists(t *testing.T) {
	assert.NotNil(t, deployCmd)
	assert.Equal(t, "deploy", deployCmd.Use)
}

func TestDeployCommand_Flags(t *testing.T) {
	flags := map[string]string{
		"cloud":      "",
		"namespace":  "kapok",
		"domain":     "kapok.local",
		"image-tag":  "latest",
		"output-dir": "",
		"context":    "",
		"timeout":    "10m",
	}

	for name, defVal := range flags {
		t.Run(name, func(t *testing.T) {
			f := deployCmd.Flags().Lookup(name)
			require.NotNil(t, f, "flag %s should exist", name)
			assert.Equal(t, defVal, f.DefValue)
		})
	}
}

func TestDeployCommand_BoolFlags(t *testing.T) {
	boolFlags := map[string]string{
		"tls":     "false",
		"hpa":     "true",
		"keda":    "false",
		"dry-run": "false",
	}

	for name, defVal := range boolFlags {
		t.Run(name, func(t *testing.T) {
			f := deployCmd.Flags().Lookup(name)
			require.NotNil(t, f, "flag %s should exist", name)
			assert.Equal(t, defVal, f.DefValue)
		})
	}
}

func TestDeployCommand_Help(t *testing.T) {
	assert.Contains(t, deployCmd.Long, "AWS EKS")
	assert.Contains(t, deployCmd.Long, "GCP GKE")
	assert.Contains(t, deployCmd.Long, "Azure AKS")
	assert.Contains(t, deployCmd.Long, "--dry-run")
}
