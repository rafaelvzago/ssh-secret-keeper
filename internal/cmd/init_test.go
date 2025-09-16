package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rzago/ssh-secret-keeper/internal/config"
)

func TestNewInitCommand(t *testing.T) {
	cfg := config.Default()
	cmd := newInitCommand(cfg)

	if cmd == nil {
		t.Fatal("newInitCommand() returned nil")
	}

	if cmd.Use != "init" {
		t.Errorf("Expected command use 'init', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Init command short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Init command long description is empty")
	}
}

func TestInitCommand_Flags(t *testing.T) {
	cfg := config.Default()
	cmd := newInitCommand(cfg)

	expectedFlags := []string{"vault-addr", "token", "mount-path", "config-path", "force"}

	for _, flagName := range expectedFlags {
		if cmd.Flag(flagName) == nil {
			t.Errorf("Expected --%s flag to be present", flagName)
		}
	}
}

func TestInitCommand_Help(t *testing.T) {
	cfg := config.Default()
	cmd := newInitCommand(cfg)
	cmd.SetArgs([]string{"--help"})

	// Capture output
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Init help command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Initialize SSH Secret Keeper") {
		t.Error("Help output doesn't contain init description")
	}

	if !strings.Contains(output, "--vault-addr") {
		t.Error("Help output doesn't show --vault-addr flag")
	}

	if !strings.Contains(output, "--token") {
		t.Error("Help output doesn't show --token flag")
	}
}

func TestInitCommand_FlagParsing(t *testing.T) {
	cfg := config.Default()
	cmd := newInitCommand(cfg)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		checkFn func() bool
	}{
		{
			name:    "vault-addr flag",
			args:    []string{"--vault-addr", "https://vault.example.com:8200"},
			wantErr: false,
			checkFn: func() bool {
				vaultAddr, _ := cmd.Flags().GetString("vault-addr")
				return vaultAddr == "https://vault.example.com:8200"
			},
		},
		{
			name:    "token flag",
			args:    []string{"--token", "hvs.test-token"},
			wantErr: false,
			checkFn: func() bool {
				token, _ := cmd.Flags().GetString("token")
				return token == "hvs.test-token"
			},
		},
		{
			name:    "mount-path flag",
			args:    []string{"--mount-path", "custom-ssh-backups"},
			wantErr: false,
			checkFn: func() bool {
				mountPath, _ := cmd.Flags().GetString("mount-path")
				return mountPath == "custom-ssh-backups"
			},
		},
		{
			name:    "config-path flag",
			args:    []string{"--config-path", "/custom/config.yaml"},
			wantErr: false,
			checkFn: func() bool {
				configPath, _ := cmd.Flags().GetString("config-path")
				return configPath == "/custom/config.yaml"
			},
		},
		{
			name:    "force flag",
			args:    []string{"--force"},
			wantErr: false,
			checkFn: func() bool {
				force, _ := cmd.Flags().GetBool("force")
				return force
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.ParseFlags(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFn != nil && !tt.checkFn() {
				t.Error("Flag parsing check failed")
			}
		})
	}
}

func TestInitCommand_DefaultValues(t *testing.T) {
	cfg := config.Default()
	cmd := newInitCommand(cfg)

	// Check default flag values
	vaultAddr, _ := cmd.Flags().GetString("vault-addr")
	if vaultAddr == "" {
		t.Error("Expected vault-addr to have default value")
	}

	token, _ := cmd.Flags().GetString("token")
	if token != "" {
		t.Errorf("Expected empty default token, got '%s'", token)
	}

	mountPath, _ := cmd.Flags().GetString("mount-path")
	if mountPath != "ssh-backups" {
		t.Errorf("Expected default mount-path 'ssh-backups', got '%s'", mountPath)
	}

	force, _ := cmd.Flags().GetBool("force")
	if force {
		t.Error("Expected force to default to false")
	}
}

func TestInitCommand_Validation(t *testing.T) {
	cfg := config.Default()
	cmd := newInitCommand(cfg)

	// Test that the command structure is valid
	if cmd.RunE == nil {
		t.Error("Init command should have a RunE function")
	}

	// Test that the command has proper structure
	if cmd.Short == "" {
		t.Error("Init command should have a short description")
	}
}
