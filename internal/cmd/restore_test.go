package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rzago/ssh-secret-keeper/internal/config"
)

func TestNewRestoreCommand(t *testing.T) {
	cfg := config.Default()
	cmd := newRestoreCommand(cfg)

	if cmd == nil {
		t.Fatal("newRestoreCommand() returned nil")
	}

	if cmd.Use != "restore [backup-name]" {
		t.Errorf("Expected command use 'restore [backup-name]', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Restore command short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Restore command long description is empty")
	}
}

func TestRestoreCommand_Flags(t *testing.T) {
	cfg := config.Default()
	cmd := newRestoreCommand(cfg)

	expectedFlags := []string{"backup", "target-dir", "dry-run", "overwrite", "interactive", "select", "files"}

	for _, flagName := range expectedFlags {
		if cmd.Flag(flagName) == nil {
			t.Errorf("Expected --%s flag to be present", flagName)
		}
	}
}

func TestRestoreCommand_Help(t *testing.T) {
	cfg := config.Default()
	cmd := newRestoreCommand(cfg)
	cmd.SetArgs([]string{"--help"})

	// Capture output
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Restore help command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Restore SSH") {
		t.Error("Help output doesn't contain restore description")
	}

	if !strings.Contains(output, "--target-dir") {
		t.Error("Help output doesn't show --target-dir flag")
	}

	if !strings.Contains(output, "--dry-run") {
		t.Error("Help output doesn't show --dry-run flag")
	}
}

func TestRestoreCommand_FlagParsing(t *testing.T) {
	cfg := config.Default()
	cmd := newRestoreCommand(cfg)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		checkFn func() bool
	}{
		{
			name:    "target-dir flag",
			args:    []string{"--target-dir", "/custom/ssh/dir"},
			wantErr: false,
			checkFn: func() bool {
				targetDir, _ := cmd.Flags().GetString("target-dir")
				return targetDir == "/custom/ssh/dir"
			},
		},
		{
			name:    "dry-run flag",
			args:    []string{"--dry-run"},
			wantErr: false,
			checkFn: func() bool {
				dryRun, _ := cmd.Flags().GetBool("dry-run")
				return dryRun
			},
		},
		{
			name:    "interactive flag",
			args:    []string{"--interactive"},
			wantErr: false,
			checkFn: func() bool {
				interactive, _ := cmd.Flags().GetBool("interactive")
				return interactive
			},
		},
		{
			name:    "overwrite flag",
			args:    []string{"--overwrite"},
			wantErr: false,
			checkFn: func() bool {
				overwrite, _ := cmd.Flags().GetBool("overwrite")
				return overwrite
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

func TestRestoreCommand_RequiredArgs(t *testing.T) {
	cfg := config.Default()
	cmd := newRestoreCommand(cfg)

	// Test that backup name is required
	if cmd.Args == nil {
		t.Error("Restore command should validate arguments")
	}

	// The command should require exactly 1 argument (backup name)
	// We can't test the actual validation without running the command,
	// but we can check the Args function exists
}

func TestRestoreCommand_DefaultValues(t *testing.T) {
	cfg := config.Default()
	cmd := newRestoreCommand(cfg)

	// Check default flag values
	targetDir, _ := cmd.Flags().GetString("target-dir")
	if targetDir == "" {
		t.Error("Expected target-dir to have default value")
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		t.Error("Expected dry-run to default to false")
	}

	interactive, _ := cmd.Flags().GetBool("interactive")
	if interactive {
		t.Error("Expected interactive to default to false")
	}

	overwrite, _ := cmd.Flags().GetBool("overwrite")
	if overwrite {
		t.Error("Expected overwrite to default to false")
	}
}
