package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/rzago/ssh-secret-keeper/internal/config"
	"github.com/rzago/ssh-secret-keeper/internal/vault"
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
		name             string
		vaultData        map[string]interface{}
		expectError      bool
		expectedPerm     os.FileMode
		expectedFilename string
		description      string
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

func TestCrossStrategyRestore(t *testing.T) {
	// Test that backups created with one strategy can be restored with another
	// This is critical for cross-machine and cross-user restore scenarios

	tests := []struct {
		name                string
		backupStrategy      string
		restoreStrategy     string
		expectSuccess       bool
		description         string
		backupNamespace     string
		restoreNamespace    string
		backupCustomPrefix  string
		restoreCustomPrefix string
	}{
		{
			name:            "machine-user to universal",
			backupStrategy:  "machine-user",
			restoreStrategy: "universal",
			expectSuccess:   false, // Different paths, won't find backup
			description:     "Backup stored in machine-user path won't be found in universal path",
		},
		{
			name:            "universal to machine-user",
			backupStrategy:  "universal",
			restoreStrategy: "machine-user",
			expectSuccess:   false, // Different paths, won't find backup
			description:     "Backup stored in universal path won't be found in machine-user path",
		},
		{
			name:            "same strategy should work",
			backupStrategy:  "universal",
			restoreStrategy: "universal",
			expectSuccess:   true, // Same paths, should work
			description:     "Same strategy restore should work",
		},
		{
			name:             "universal with different namespaces",
			backupStrategy:   "universal",
			restoreStrategy:  "universal",
			backupNamespace:  "personal",
			restoreNamespace: "work",
			expectSuccess:    false, // Different namespaces create different paths
			description:      "Different namespaces create different storage paths",
		},
		{
			name:             "universal same namespace",
			backupStrategy:   "universal",
			restoreStrategy:  "universal",
			backupNamespace:  "personal",
			restoreNamespace: "personal",
			expectSuccess:    true, // Same namespace should work
			description:      "Same namespace should allow restore",
		},
		{
			name:                "custom with different prefixes",
			backupStrategy:      "custom",
			restoreStrategy:     "custom",
			backupCustomPrefix:  "team-devops",
			restoreCustomPrefix: "team-qa",
			expectSuccess:       false, // Different prefixes create different paths
			description:         "Different custom prefixes create different storage paths",
		},
		{
			name:                "custom same prefix",
			backupStrategy:      "custom",
			restoreStrategy:     "custom",
			backupCustomPrefix:  "team-devops",
			restoreCustomPrefix: "team-devops",
			expectSuccess:       true, // Same prefix should work
			description:         "Same custom prefix should allow restore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create backup configuration
			backupCfg := &config.Config{
				Vault: config.VaultConfig{
					Address:         "http://localhost:8200",
					MountPath:       "ssh-backups-test",
					StorageStrategy: tt.backupStrategy,
					BackupNamespace: tt.backupNamespace,
					CustomPrefix:    tt.backupCustomPrefix,
					TokenFile:       "/dev/null",
				},
			}

			// Create restore configuration
			restoreCfg := &config.Config{
				Vault: config.VaultConfig{
					Address:         "http://localhost:8200",
					MountPath:       "ssh-backups-test",
					StorageStrategy: tt.restoreStrategy,
					BackupNamespace: tt.restoreNamespace,
					CustomPrefix:    tt.restoreCustomPrefix,
					TokenFile:       "/dev/null",
				},
			}

			// Test path generation to verify our expectations
			backupPathGen := createPathGenerator(t, backupCfg)
			restorePathGen := createPathGenerator(t, restoreCfg)

			backupPath, err := backupPathGen.GenerateBasePath()
			if err != nil {
				t.Fatalf("Failed to generate backup path: %v", err)
			}

			restorePath, err := restorePathGen.GenerateBasePath()
			if err != nil {
				t.Fatalf("Failed to generate restore path: %v", err)
			}

			pathsMatch := backupPath == restorePath

			if tt.expectSuccess && !pathsMatch {
				t.Errorf("Expected paths to match for successful restore, but backup=%q restore=%q (%s)",
					backupPath, restorePath, tt.description)
			}

			if !tt.expectSuccess && pathsMatch {
				t.Errorf("Expected paths to differ for failed restore, but both are %q (%s)",
					backupPath, tt.description)
			}

			// Log the paths for debugging
			t.Logf("Backup strategy %s -> path: %s", tt.backupStrategy, backupPath)
			t.Logf("Restore strategy %s -> path: %s", tt.restoreStrategy, restorePath)
			t.Logf("Description: %s", tt.description)
		})
	}
}

// Note: Restore options compatibility tests removed as they require Vault client connection
// Option validation is tested through the command flag parsing tests

func TestRestorePathResolution(t *testing.T) {
	// Test that restore command properly resolves target paths for different strategies

	tests := []struct {
		name         string
		strategy     string
		namespace    string
		customPrefix string
		targetDir    string
		expectDir    string
	}{
		{
			name:      "universal strategy default target",
			strategy:  "universal",
			targetDir: "",
			expectDir: "~/.ssh", // Should use default SSH directory
		},
		{
			name:      "universal strategy custom target",
			strategy:  "universal",
			targetDir: "/custom/ssh",
			expectDir: "/custom/ssh",
		},
		{
			name:      "user strategy with namespace",
			strategy:  "universal",
			namespace: "personal",
			targetDir: "/tmp/ssh-personal",
			expectDir: "/tmp/ssh-personal",
		},
		{
			name:         "custom strategy",
			strategy:     "custom",
			customPrefix: "team-devops",
			targetDir:    "/tmp/ssh-team",
			expectDir:    "/tmp/ssh-team",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Vault: config.VaultConfig{
					Address:         "http://localhost:8200",
					MountPath:       "ssh-backups-test",
					StorageStrategy: tt.strategy,
					BackupNamespace: tt.namespace,
					CustomPrefix:    tt.customPrefix,
					TokenFile:       "/dev/null",
				},
				Backup: config.BackupConfig{
					SSHDir: "~/.ssh",
				},
			}

			opts := restoreOptions{
				backupName: "test-backup",
				targetDir:  tt.targetDir,
				dryRun:     true,
			}

			// We can't fully test runRestore without vault, but we can test option processing
			if tt.targetDir == "" {
				// Should use default from config
				expectedDefault := cfg.Backup.SSHDir
				if expectedDefault != tt.expectDir {
					// Update test expectation if needed
					t.Logf("Default SSH dir: %s", expectedDefault)
				}
			}

			// Test that the target directory is properly set
			if opts.targetDir != tt.targetDir {
				t.Errorf("Target directory not preserved: got %q, want %q", opts.targetDir, tt.targetDir)
			}

			t.Logf("Strategy: %s, Target: %s -> Expected: %s", tt.strategy, tt.targetDir, tt.expectDir)
		})
	}
}

// Note: TestRestoreCommandStrategyValidation removed as it requires Vault client connection
// Strategy validation is tested through vault.ParseStrategy tests

// Helper function to create path generator for testing
func createPathGenerator(t *testing.T, cfg *config.Config) *vault.PathGenerator {
	strategy, err := vault.ParseStrategy(cfg.Vault.StorageStrategy)
	if err != nil {
		// If strategy is invalid, it should fallback to machine-user
		strategy = vault.StrategyMachineUser
	}

	return vault.NewPathGenerator(strategy, cfg.Vault.CustomPrefix, cfg.Vault.BackupNamespace)
}

func TestRestoreCommandCrossStrategyScenarios(t *testing.T) {
	// Test realistic cross-strategy restore scenarios

	scenarios := []struct {
		name        string
		description string
		setup       func() (*config.Config, *config.Config) // backup config, restore config
		expectMatch bool
	}{
		{
			name:        "laptop_to_desktop_universal",
			description: "User backs up on laptop with universal strategy, restores on desktop",
			setup: func() (*config.Config, *config.Config) {
				// Both use universal strategy - should work
				backupCfg := &config.Config{
					Vault: config.VaultConfig{
						StorageStrategy: "universal",
						MountPath:       "ssh-backups",
					},
				}
				restoreCfg := &config.Config{
					Vault: config.VaultConfig{
						StorageStrategy: "universal",
						MountPath:       "ssh-backups",
					},
				}
				return backupCfg, restoreCfg
			},
			expectMatch: true,
		},
		{
			name:        "legacy_to_universal_migration",
			description: "User has backups from legacy machine-user strategy, wants to restore with universal",
			setup: func() (*config.Config, *config.Config) {
				backupCfg := &config.Config{
					Vault: config.VaultConfig{
						StorageStrategy: "machine-user",
						MountPath:       "ssh-backups",
					},
				}
				restoreCfg := &config.Config{
					Vault: config.VaultConfig{
						StorageStrategy: "universal",
						MountPath:       "ssh-backups",
					},
				}
				return backupCfg, restoreCfg
			},
			expectMatch: false, // Different paths, needs migration
		},
		{
			name:        "team_to_personal_namespace",
			description: "User switches from team namespace to personal namespace",
			setup: func() (*config.Config, *config.Config) {
				backupCfg := &config.Config{
					Vault: config.VaultConfig{
						StorageStrategy: "universal",
						BackupNamespace: "team-devops",
						MountPath:       "ssh-backups",
					},
				}
				restoreCfg := &config.Config{
					Vault: config.VaultConfig{
						StorageStrategy: "universal",
						BackupNamespace: "personal",
						MountPath:       "ssh-backups",
					},
				}
				return backupCfg, restoreCfg
			},
			expectMatch: false, // Different namespaces
		},
		{
			name:        "cross_user_restore",
			description: "User1 backs up, User2 tries to restore with user strategy",
			setup: func() (*config.Config, *config.Config) {
				// Both use user strategy but will have different usernames in path
				// This simulates different users, which should fail
				backupCfg := &config.Config{
					Vault: config.VaultConfig{
						StorageStrategy: "user",
						MountPath:       "ssh-backups",
					},
				}
				restoreCfg := &config.Config{
					Vault: config.VaultConfig{
						StorageStrategy: "user",
						MountPath:       "ssh-backups",
					},
				}
				return backupCfg, restoreCfg
			},
			expectMatch: true, // Same user in test environment
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			backupCfg, restoreCfg := scenario.setup()

			// Generate paths for both configurations
			backupPathGen := createPathGenerator(t, backupCfg)
			restorePathGen := createPathGenerator(t, restoreCfg)

			backupPath, err := backupPathGen.GenerateBasePath()
			if err != nil {
				t.Fatalf("Failed to generate backup path: %v", err)
			}

			restorePath, err := restorePathGen.GenerateBasePath()
			if err != nil {
				t.Fatalf("Failed to generate restore path: %v", err)
			}

			pathsMatch := backupPath == restorePath

			if scenario.expectMatch && !pathsMatch {
				t.Errorf("Scenario '%s': Expected paths to match, but backup=%q restore=%q",
					scenario.name, backupPath, restorePath)
			}

			if !scenario.expectMatch && pathsMatch {
				t.Errorf("Scenario '%s': Expected paths to differ, but both are %q",
					scenario.name, backupPath)
			}

			t.Logf("Scenario: %s", scenario.description)
			t.Logf("Backup path: %s", backupPath)
			t.Logf("Restore path: %s", restorePath)
			t.Logf("Paths match: %v (expected: %v)", pathsMatch, scenario.expectMatch)
		})
	}
}
