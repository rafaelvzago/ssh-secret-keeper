package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rzago/ssh-secret-keeper/internal/config"
	"github.com/spf13/cobra"
)

func TestNewRootCommand(t *testing.T) {
	cfg := config.Default()
	cmd := NewRootCommand(cfg)

	if cmd == nil {
		t.Fatal("NewRootCommand() returned nil")
	}

	if cmd.Use != "sshsk" {
		t.Errorf("Expected command use 'sshsk', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Command long description is empty")
	}
}

func TestRootCommand_Flags(t *testing.T) {
	cfg := config.Default()
	cmd := NewRootCommand(cfg)

	// Test that required flags are present
	if cmd.Flag("verbose") == nil {
		t.Error("Expected --verbose flag to be present")
	}

	if cmd.Flag("quiet") == nil {
		t.Error("Expected --quiet flag to be present")
	}

	if cmd.Flag("config") == nil {
		t.Error("Expected --config flag to be present")
	}
}

func TestRootCommand_Subcommands(t *testing.T) {
	cfg := config.Default()
	cmd := NewRootCommand(cfg)

	expectedCommands := []string{
		"init", "backup", "restore", "list", "delete", "analyze", "status", "version",
	}

	for _, expectedCmd := range expectedCommands {
		found := false
		for _, subCmd := range cmd.Commands() {
			if subCmd.Name() == expectedCmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found", expectedCmd)
		}
	}
}

func TestRootCommand_Help(t *testing.T) {
	cfg := config.Default()
	cmd := NewRootCommand(cfg)
	cmd.SetArgs([]string{"--help"})

	// Capture output
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "SSH Secret Keeper") {
		t.Error("Help output doesn't contain project name")
	}

	if !strings.Contains(output, "Available Commands:") {
		t.Error("Help output doesn't show available commands")
	}
}

func TestVersionCommand(t *testing.T) {
	cmd := newVersionCommand()

	if cmd == nil {
		t.Fatal("newVersionCommand() returned nil")
	}

	if cmd.Use != "version" {
		t.Errorf("Expected command use 'version', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Version command short description is empty")
	}

	// Test version command execution
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	if cmd.Run != nil {
		cmd.Run(cmd, []string{})
	} else {
		t.Error("Version command should have a Run function")
	}

	output := buf.String()
	if !strings.Contains(output, "SSH Secret Keeper") {
		t.Error("Version output doesn't contain project name")
	}
}

func TestSetupLogging(t *testing.T) {
	cfg := config.Default()
	cmd := NewRootCommand(cfg)

	// Test verbose flag
	cmd.SetArgs([]string{"--verbose"})
	err := cmd.ParseFlags([]string{"--verbose"})
	if err != nil {
		t.Fatalf("Failed to parse verbose flag: %v", err)
	}

	// setupLogging should not panic
	setupLogging(cmd)

	// Test quiet flag
	cmd.ResetFlags()
	cmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging")
	cmd.PersistentFlags().Bool("quiet", false, "Suppress all output except errors")
	cmd.SetArgs([]string{"--quiet"})
	err = cmd.ParseFlags([]string{"--quiet"})
	if err != nil {
		t.Fatalf("Failed to parse quiet flag: %v", err)
	}

	// setupLogging should not panic
	setupLogging(cmd)
}

func TestHandleError(t *testing.T) {
	// Test with nil error (should not exit)
	handleError(nil, "test message")

	// Note: We can't easily test the error case since it calls os.Exit(1)
	// In a real scenario, we'd refactor handleError to be more testable
}

func TestRootCommand_FlagParsing(t *testing.T) {
	cfg := config.Default()
	cmd := NewRootCommand(cfg)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		checkFn func(*cobra.Command) bool
	}{
		{
			name:    "verbose flag",
			args:    []string{"--verbose"},
			wantErr: false,
			checkFn: func(cmd *cobra.Command) bool {
				verbose, _ := cmd.Flags().GetBool("verbose")
				return verbose
			},
		},
		{
			name:    "quiet flag",
			args:    []string{"--quiet"},
			wantErr: false,
			checkFn: func(cmd *cobra.Command) bool {
				quiet, _ := cmd.Flags().GetBool("quiet")
				return quiet
			},
		},
		{
			name:    "config flag",
			args:    []string{"--config", "/path/to/config.yaml"},
			wantErr: false,
			checkFn: func(cmd *cobra.Command) bool {
				configPath, _ := cmd.Flags().GetString("config")
				return configPath == "/path/to/config.yaml"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd.SetArgs(tt.args)
			err := cmd.ParseFlags(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFn != nil && !tt.checkFn(cmd) {
				t.Error("Flag parsing check failed")
			}
		})
	}
}
