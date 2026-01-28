package deploy

import (
	"bytes"
	"fmt"
	"os/exec"
)

// CommandRunner abstracts command execution for testability.
type CommandRunner interface {
	Run(name string, args ...string) (string, error)
}

// ExecRunner runs commands via os/exec.
type ExecRunner struct{}

// Run executes a command and returns combined output.
func (r *ExecRunner) Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s failed: %s: %w", name, stderr.String(), err)
	}
	return stdout.String(), nil
}
