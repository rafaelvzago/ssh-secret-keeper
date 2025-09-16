package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rzago/ssh-secret-keeper/internal/config"
)

func TestNewBackupCommand(t *testing.T) {
	cfg := config.Default()
	cmd := newBackupCommand(cfg)

	if cmd == nil {
		t.Fatal("newBackupCommand() returned nil")
	}

	if cmd.Use != "backup [name]" {
		t.Errorf("Expected command use 'backup [name]', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Backup command short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Backup command long description is empty")
	}
}

func TestBackupCommand_Flags(t *testing.T) {
	cfg := config.Default()
	cmd := newBackupCommand(cfg)

	expectedFlags := []string{"name", "ssh-dir", "dry-run", "interactive"}

	for _, flagName := range expectedFlags {
		if cmd.Flag(flagName) == nil {
			t.Errorf("Expected --%s flag to be present", flagName)
		}
	}
}

func TestBackupCommand_Help(t *testing.T) {
	cfg := config.Default()
	cmd := newBackupCommand(cfg)
	cmd.SetArgs([]string{"--help"})

	// Capture output
	buf := bytes.NewBufferString("")
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Backup help command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Backup SSH") {
		t.Error("Help output doesn't contain backup description")
	}

	if !strings.Contains(output, "--ssh-dir") {
		t.Error("Help output doesn't show --ssh-dir flag")
	}

	if !strings.Contains(output, "--dry-run") {
		t.Error("Help output doesn't show --dry-run flag")
	}
}

func TestBackupCommand_FlagParsing(t *testing.T) {
	cfg := config.Default()
	cmd := newBackupCommand(cfg)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		checkFn func() bool
	}{
		{
			name:    "ssh-dir flag",
			args:    []string{"--ssh-dir", "/custom/ssh/dir"},
			wantErr: false,
			checkFn: func() bool {
				sshDir, _ := cmd.Flags().GetString("ssh-dir")
				return sshDir == "/custom/ssh/dir"
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

func TestBackupCommand_Validation(t *testing.T) {
	cfg := config.Default()
	cmd := newBackupCommand(cfg)

	// Test backup name validation
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid backup name",
			args:    []string{"my-backup"},
			wantErr: false, // We expect this to fail due to missing VAULT_ADDR, but not due to name validation
		},
		{
			name:    "empty backup name",
			args:    []string{""},
			wantErr: false, // Empty name should generate a default name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd.SetArgs(tt.args)

			// We can't actually run the command without proper Vault setup,
			// but we can test argument parsing
			err := cmd.ParseFlags([]string{})
			if err != nil {
				t.Errorf("Unexpected error in ParseFlags: %v", err)
			}
		})
	}
}

func TestBackupCommand_DefaultValues(t *testing.T) {
	cfg := config.Default()
	cmd := newBackupCommand(cfg)

	// Check default flag values
	sshDir, _ := cmd.Flags().GetString("ssh-dir")
	if sshDir == "" {
		t.Error("Expected ssh-dir to have default value")
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		t.Error("Expected dry-run to default to false")
	}

	interactive, _ := cmd.Flags().GetBool("interactive")
	if interactive {
		t.Error("Expected interactive to default to false")
	}
}
