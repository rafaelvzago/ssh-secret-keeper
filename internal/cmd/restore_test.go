package cmd

import (
	"bytes"
	"os"
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

func TestParseVaultBackup_MissingPermissions(t *testing.T) {
	tests := []struct {
		name              string
		vaultData         map[string]interface{}
		expectError       bool
		expectedPerm      os.FileMode
		expectedFilename  string
		description       string
	}{
		{
			name: "missing permissions field",
			vaultData: map[string]interface{}{
				"version":   "1.0",
				"hostname":  "test-host",
				"username":  "test-user",
				"ssh_dir":   "/home/test/.ssh",
				"timestamp": "2023-01-01T00:00:00Z",
				"files": map[string]interface{}{
					"id_rsa": map[string]interface{}{
						"size":     float64(1024),
						"mod_time": "2023-01-01T00:00:00Z",
						"checksum": "abc123",
						// permissions field is missing
					},
				},
			},
			expectError:      false,
			expectedPerm:     0600,
			expectedFilename: "id_rsa",
			description:      "Should use 0600 fallback when permissions field is missing",
		},
		{
			name: "nil permissions value",
			vaultData: map[string]interface{}{
				"version":   "1.0",
				"hostname":  "test-host",
				"username":  "test-user",
				"ssh_dir":   "/home/test/.ssh",
				"timestamp": "2023-01-01T00:00:00Z",
				"files": map[string]interface{}{
					"id_rsa.pub": map[string]interface{}{
						"permissions": nil,
						"size":        float64(512),
						"mod_time":    "2023-01-01T00:00:00Z",
						"checksum":    "def456",
					},
				},
			},
			expectError:      false,
			expectedPerm:     0600,
			expectedFilename: "id_rsa.pub",
			description:      "Should use 0600 fallback when permissions is nil",
		},
		{
			name: "string permissions value",
			vaultData: map[string]interface{}{
				"version":   "1.0",
				"hostname":  "test-host",
				"username":  "test-user",
				"ssh_dir":   "/home/test/.ssh",
				"timestamp": "2023-01-01T00:00:00Z",
				"files": map[string]interface{}{
					"config": map[string]interface{}{
						"permissions": "0644", // string instead of number
						"size":        float64(256),
						"mod_time":    "2023-01-01T00:00:00Z",
						"checksum":    "ghi789",
					},
				},
			},
			expectError:      false,
			expectedPerm:     0600,
			expectedFilename: "config",
			description:      "Should use 0600 fallback when permissions is a string",
		},
		{
			name: "boolean permissions value",
			vaultData: map[string]interface{}{
				"version":   "1.0",
				"hostname":  "test-host",
				"username":  "test-user",
				"ssh_dir":   "/home/test/.ssh",
				"timestamp": "2023-01-01T00:00:00Z",
				"files": map[string]interface{}{
					"known_hosts": map[string]interface{}{
						"permissions": true, // boolean instead of number
						"size":        float64(128),
						"mod_time":    "2023-01-01T00:00:00Z",
						"checksum":    "jkl012",
					},
				},
			},
			expectError:      false,
			expectedPerm:     0600,
			expectedFilename: "known_hosts",
			description:      "Should use 0600 fallback when permissions is a boolean",
		},
		{
			name: "valid float64 permissions should work normally",
			vaultData: map[string]interface{}{
				"version":   "1.0",
				"hostname":  "test-host",
				"username":  "test-user",
				"ssh_dir":   "/home/test/.ssh",
				"timestamp": "2023-01-01T00:00:00Z",
				"files": map[string]interface{}{
					"id_rsa": map[string]interface{}{
						"permissions": float64(0644),
						"size":        float64(1024),
						"mod_time":    "2023-01-01T00:00:00Z",
						"checksum":    "mno345",
					},
				},
			},
			expectError:      false,
			expectedPerm:     0644,
			expectedFilename: "id_rsa",
			description:      "Should preserve valid float64 permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backup, err := parseVaultBackup(tt.vaultData)
			
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none: %s", tt.description)
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v (%s)", err, tt.description)
				return
			}
			
			if !tt.expectError && backup != nil {
				fileData, exists := backup.Files[tt.expectedFilename]
				if !exists {
					t.Errorf("Expected file %s not found in backup (%s)", tt.expectedFilename, tt.description)
					return
				}
				
				actualPerm := fileData.Permissions & os.ModePerm
				if actualPerm != tt.expectedPerm {
					t.Errorf("Expected permissions %04o, got %04o (%s)", 
						tt.expectedPerm, actualPerm, tt.description)
				}
				
				// Verify other fields are preserved
				if fileData.Filename != tt.expectedFilename {
					t.Errorf("Expected filename %s, got %s", tt.expectedFilename, fileData.Filename)
				}
			}
		})
	}
}

func TestParseVaultBackup_MultipleFilesWithMixedPermissions(t *testing.T) {
	// Test scenario where some files have valid permissions and others don't
	vaultData := map[string]interface{}{
		"version":   "1.0",
		"hostname":  "test-host",
		"username":  "test-user",
		"ssh_dir":   "/home/test/.ssh",
		"timestamp": "2023-01-01T00:00:00Z",
		"files": map[string]interface{}{
			"id_rsa": map[string]interface{}{
				"permissions": float64(0600), // valid
				"size":        float64(1024),
				"mod_time":    "2023-01-01T00:00:00Z",
				"checksum":    "abc123",
			},
			"id_rsa.pub": map[string]interface{}{
				"permissions": "invalid", // invalid - should use fallback
				"size":        float64(512),
				"mod_time":    "2023-01-01T00:00:00Z",
				"checksum":    "def456",
			},
			"config": map[string]interface{}{
				// missing permissions - should use fallback
				"size":     float64(256),
				"mod_time": "2023-01-01T00:00:00Z",
				"checksum": "ghi789",
			},
		},
	}

	backup, err := parseVaultBackup(vaultData)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that we have all files
	if len(backup.Files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(backup.Files))
	}

	// Check valid permissions are preserved
	if idRsa, exists := backup.Files["id_rsa"]; exists {
		actualPerm := idRsa.Permissions & os.ModePerm
		if actualPerm != 0600 {
			t.Errorf("Expected id_rsa permissions 0600, got %04o", actualPerm)
		}
	} else {
		t.Error("id_rsa file missing")
	}

	// Check fallback permissions are applied
	if idRsaPub, exists := backup.Files["id_rsa.pub"]; exists {
		actualPerm := idRsaPub.Permissions & os.ModePerm
		if actualPerm != 0600 {
			t.Errorf("Expected id_rsa.pub fallback permissions 0600, got %04o", actualPerm)
		}
	} else {
		t.Error("id_rsa.pub file missing")
	}

	if config, exists := backup.Files["config"]; exists {
		actualPerm := config.Permissions & os.ModePerm
		if actualPerm != 0600 {
			t.Errorf("Expected config fallback permissions 0600, got %04o", actualPerm)
		}
	} else {
		t.Error("config file missing")
	}
}
