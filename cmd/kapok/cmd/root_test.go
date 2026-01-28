package cmd_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

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
	if testing.Short() {
		t.Skip("Skipping dev command test in short mode - requires Docker")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	buf := new(bytes.Buffer)
	err := cmd.ExecuteWithContext(ctx, buf, []string{"dev"})

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

func TestTenantCommands_Help(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantOut string
	}{
		{
			name:    "tenant help",
			args:    []string{"tenant", "--help"},
			wantOut: "tenant",
		},
		{
			name:    "tenant create help",
			args:    []string{"tenant", "create", "--help"},
			wantOut: "create",
		},
		{
			name:    "tenant list help",
			args:    []string{"tenant", "list", "--help"},
			wantOut: "list",
		},
		{
			name:    "tenant delete help",
			args:    []string{"tenant", "delete", "--help"},
			wantOut: "delete",
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

func TestTenantCommands_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tenant integration tests - requires database")
	}

	tests := []struct {
		name    string
		args    []string
		wantOut string
	}{
		{
			name:    "tenant create",
			args:    []string{"tenant", "create", "test"},
			wantOut: "Tenant created successfully",
		},
		{
			name:    "tenant list",
			args:    []string{"tenant", "list"},
			wantOut: "ID",
		},
		{
			name:    "tenant delete",
			args:    []string{"tenant", "delete", "--force", "tenant_123"},
			wantOut: "deleted",
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
