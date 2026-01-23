package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kapok/kapok/cmd/kapok/cmd"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantOut  string
		wantErr  bool
	}{
		{
			name:    "help flag",
			args:    []string{"--help"},
			wantOut: "Backend-as-a-Service",
			wantErr: false,
		},
		{
			name:    "version command",
			args:    []string{"version"},
			wantOut: "kapok version dev",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := cmd.ExecuteContext(buf, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := buf.String()
			if !strings.Contains(got, tt.wantOut) {
				t.Errorf("ExecuteContext() got = %v, want substring %v", got, tt.wantOut)
			}
		})
	}
}

func TestInitCommand(t *testing.T) {
	buf := new(bytes.Buffer)
	err := cmd.ExecuteContext(buf, []string{"init", "test-project"})

	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "test-project") {
		t.Errorf("Expected output to contain 'test-project', got: %v", got)
	}
	if !strings.Contains(got, "Created kapok.yaml") {
		t.Errorf("Expected output to contain placeholder confirmation")
	}
}

func TestDevCommand(t *testing.T) {
	buf := new(bytes.Buffer)
	err := cmd.ExecuteContext(buf, []string{"dev"})

	if err != nil {
		t.Fatalf("dev command failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "development environment") {
		t.Errorf("Expected output to contain development environment info")
	}
}

func TestDeployCommand(t *testing.T) {
	buf := new(bytes.Buffer)
	err := cmd.ExecuteContext(buf, []string{"deploy"})

	if err != nil {
		t.Fatalf("deploy command failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "Deploying Kapok") {
		t.Errorf("Expected output to contain deployment info")
	}
}

func TestTenantCommands(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantOut string
	}{
		{
			name:    "tenant create",
			args:   []string{"tenant", "create", "--name=test"},
			wantOut: "Creating tenant: test",
		},
		{
			name:    "tenant list",
			args:    []string{"tenant", "list"},
			wantOut: "ID",
		},
		{
			name:    "tenant delete",
			args:    []string{"tenant", "delete", "tenant_123"},
			wantOut: "Deleting tenant: tenant_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := cmd.ExecuteContext(buf, tt.args)

			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}

			got := buf.String()
			if !strings.Contains(got, tt.wantOut) {
				t.Errorf("Expected output to contain '%s', got: %v", tt.wantOut, got)
			}
		})
	}
}
