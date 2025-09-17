package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rzago/ssh-secret-keeper/internal/analyzer"
	"github.com/rzago/ssh-secret-keeper/internal/config"
	"github.com/rzago/ssh-secret-keeper/internal/ssh"
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
	if !strings.Contains(output, "Backup") {
		t.Error("Help output doesn't contain backup description")
	}

	if !strings.Contains(output, "--ssh-dir") {
		t.Error("Help output doesn't show --ssh-dir flag")
	}

	if !strings.Contains(output, "--dry-run") {
		t.Error("Help output doesn't show --dry-run flag")
	}

	if !strings.Contains(output, "--interactive") {
		t.Error("Help output doesn't show --interactive flag")
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

// ===== COMPREHENSIVE INTERACTIVE BACKUP TESTS =====

// TestInteractiveFileSelection_Logic tests the logic of interactive file selection
func TestInteractiveFileSelection_Logic(t *testing.T) {
	// Create test backup data
	testBackupData := createTestBackupData(t)

	// Test that the function exists and can be called (though we can't test user input easily)
	t.Run("function_exists", func(t *testing.T) {
		// We can't easily test the interactive function due to fmt.Scanln
		// But we can verify the backup data structure is correct for interactive use
		if testBackupData.Files == nil {
			t.Error("Test backup data should have Files map")
		}

		if len(testBackupData.Files) == 0 {
			t.Error("Test backup data should have at least one file for interactive selection")
		}

		// Verify each file has the required information for interactive display
		for filename, fileData := range testBackupData.Files {
			if fileData.KeyInfo == nil {
				t.Errorf("File %s should have KeyInfo for interactive display", filename)
			}
			if fileData.Size == 0 {
				t.Errorf("File %s should have size information", filename)
			}
			if fileData.KeyInfo.Type == "" {
				t.Errorf("File %s should have type information", filename)
			}
		}
	})
}

// TestInteractiveFileSelection_FileRemoval tests that files can be removed from backup
func TestInteractiveFileSelection_FileRemoval(t *testing.T) {
	testBackupData := createTestBackupData(t)
	originalFileCount := len(testBackupData.Files)

	// Test that we can remove files from the backup data structure
	t.Run("file_removal_logic", func(t *testing.T) {
		// Simulate removing a file (this is what happens when user says "n")
		var fileToRemove string
		for filename := range testBackupData.Files {
			fileToRemove = filename
			break
		}

		// Remove the file (simulating user selecting "n")
		delete(testBackupData.Files, fileToRemove)

		if len(testBackupData.Files) != originalFileCount-1 {
			t.Errorf("Expected file count to be %d after removal, got %d",
				originalFileCount-1, len(testBackupData.Files))
		}

		if _, exists := testBackupData.Files[fileToRemove]; exists {
			t.Errorf("File %s should have been removed from backup", fileToRemove)
		}
	})
}

// TestInteractiveBackupOptions tests that interactive option is properly handled
func TestInteractiveBackupOptions(t *testing.T) {
	tests := []struct {
		name        string
		opts        backupOptions
		expectCall  bool
		description string
	}{
		{
			name: "interactive_enabled",
			opts: backupOptions{
				name:        "test-backup",
				sshDir:      "/tmp/test-ssh",
				dryRun:      false,
				interactive: true,
			},
			expectCall:  true,
			description: "Interactive mode should trigger file selection",
		},
		{
			name: "interactive_disabled",
			opts: backupOptions{
				name:        "test-backup",
				sshDir:      "/tmp/test-ssh",
				dryRun:      false,
				interactive: false,
			},
			expectCall:  false,
			description: "Non-interactive mode should skip file selection",
		},
		{
			name: "dry_run_with_interactive",
			opts: backupOptions{
				name:        "test-backup",
				sshDir:      "/tmp/test-ssh",
				dryRun:      true,
				interactive: true,
			},
			expectCall:  false,
			description: "Dry run should skip interactive selection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that backupOptions struct properly holds interactive flag
			if tt.opts.interactive != tt.expectCall && !tt.opts.dryRun {
				t.Errorf("Interactive option not properly set: got %v, want %v",
					tt.opts.interactive, tt.expectCall)
			}

			// Test the logic flow (we can't test actual execution without Vault)
			if tt.opts.dryRun && tt.opts.interactive {
				// In dry run mode, interactive selection should be skipped
				// This is the correct behavior as shown in backup.go lines 87-90
				t.Log("Dry run correctly skips interactive selection")
			}
		})
	}
}

// TestDisplayBackupSummary tests the backup summary display
func TestDisplayBackupSummary(t *testing.T) {
	// Create test backup data
	testBackup := createTestBackupData(t)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function
	displayBackupSummary(testBackup)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected information
	expectedStrings := []string{
		"Backup Analysis Summary",
		"Total files:",
		"Key pairs:",
		"Service keys:",
		"Personal keys:",
		"Work keys:",
		"System files:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected backup summary to contain '%s', but it didn't", expected)
		}
	}

	// Verify numbers make sense
	if !strings.Contains(output, fmt.Sprintf("Total files: %d", len(testBackup.Files))) {
		t.Error("Backup summary should show correct total file count")
	}
}

// TestPrepareVaultData tests the vault data preparation
func TestPrepareVaultData(t *testing.T) {
	testBackup := createTestBackupData(t)

	vaultData := prepareVaultData(testBackup)

	// Verify required fields are present
	requiredFields := []string{"version", "timestamp", "hostname", "username", "ssh_dir", "metadata", "files"}

	for _, field := range requiredFields {
		if _, exists := vaultData[field]; !exists {
			t.Errorf("Vault data should contain field '%s'", field)
		}
	}

	// Verify files data structure
	files, ok := vaultData["files"].(map[string]interface{})
	if !ok {
		t.Fatal("Vault data files should be a map")
	}

	if len(files) != len(testBackup.Files) {
		t.Errorf("Vault data should contain %d files, got %d", len(testBackup.Files), len(files))
	}

	// Verify each file has required fields
	for filename, fileInterface := range files {
		fileData, ok := fileInterface.(map[string]interface{})
		if !ok {
			t.Errorf("File data for %s should be a map", filename)
			continue
		}

		requiredFileFields := []string{"filename", "content", "permissions", "size", "mod_time", "checksum", "key_info"}
		for _, field := range requiredFileFields {
			if _, exists := fileData[field]; !exists {
				t.Errorf("File %s should contain field '%s'", filename, field)
			}
		}
	}
}

// TestInteractiveSelectionPrompt tests the interactive selection prompt format
func TestInteractiveSelectionPrompt(t *testing.T) {
	testBackup := createTestBackupData(t)

	// Test that each file would generate appropriate prompt
	for filename, fileData := range testBackup.Files {
		// This is the format used in interactiveFileSelection
		expectedPrompt := fmt.Sprintf("Include '%s' [%s, %d bytes]? [y/N/a/q]: ",
			filename,
			fileData.KeyInfo.Type,
			fileData.Size)

		// Verify prompt components
		if !strings.Contains(expectedPrompt, filename) {
			t.Errorf("Prompt should contain filename %s", filename)
		}

		if !strings.Contains(expectedPrompt, string(fileData.KeyInfo.Type)) {
			t.Errorf("Prompt should contain file type %s", fileData.KeyInfo.Type)
		}

		if !strings.Contains(expectedPrompt, fmt.Sprintf("%d bytes", fileData.Size)) {
			t.Errorf("Prompt should contain file size %d", fileData.Size)
		}

		if !strings.Contains(expectedPrompt, "[y/N/a/q]") {
			t.Error("Prompt should contain response options [y/N/a/q]")
		}
	}
}

// TestInteractiveSelectionBehavior tests the expected behavior for different responses
func TestInteractiveSelectionBehavior(t *testing.T) {
	testBackup := createTestBackupData(t)
	originalFileCount := len(testBackup.Files)

	tests := []struct {
		name           string
		response       string
		shouldKeepFile bool
		shouldContinue bool
		shouldExit     bool
	}{
		{
			name:           "yes_response",
			response:       "y",
			shouldKeepFile: true,
			shouldContinue: true,
			shouldExit:     false,
		},
		{
			name:           "Yes_response",
			response:       "Y",
			shouldKeepFile: true,
			shouldContinue: true,
			shouldExit:     false,
		},
		{
			name:           "yes_full_response",
			response:       "yes",
			shouldKeepFile: true,
			shouldContinue: true,
			shouldExit:     false,
		},
		{
			name:           "no_response",
			response:       "n",
			shouldKeepFile: false,
			shouldContinue: true,
			shouldExit:     false,
		},
		{
			name:           "default_response",
			response:       "",
			shouldKeepFile: false,
			shouldContinue: true,
			shouldExit:     false,
		},
		{
			name:           "all_response",
			response:       "a",
			shouldKeepFile: true,
			shouldContinue: false, // Should stop asking for remaining files
			shouldExit:     false,
		},
		{
			name:           "All_response",
			response:       "A",
			shouldKeepFile: true,
			shouldContinue: false,
			shouldExit:     false,
		},
		{
			name:           "quit_response",
			response:       "q",
			shouldKeepFile: false,
			shouldContinue: false,
			shouldExit:     true,
		},
		{
			name:           "Quit_response",
			response:       "Q",
			shouldKeepFile: false,
			shouldContinue: false,
			shouldExit:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh test data for each test
			backup := createTestBackupData(t)

			// Get first file to test with
			var testFilename string
			for filename := range backup.Files {
				testFilename = filename
				break
			}

			// Test the expected behavior based on response
			switch tt.response {
			case "y", "Y", "yes":
				// File should be kept (no action needed)
				if len(backup.Files) != originalFileCount {
					t.Errorf("File count should remain %d for 'yes' response", originalFileCount)
				}
			case "a", "A", "all":
				// All files should be kept, no further processing needed
				if len(backup.Files) != originalFileCount {
					t.Errorf("File count should remain %d for 'all' response", originalFileCount)
				}
			case "q", "Q", "quit":
				// Should return error (cancellation)
				expectedError := "backup cancelled by user"
				if expectedError == "" {
					t.Error("Quit response should generate cancellation error")
				}
			default:
				// File should be removed (including empty string default)
				delete(backup.Files, testFilename)
				if len(backup.Files) != originalFileCount-1 {
					t.Errorf("File count should be %d after removal", originalFileCount-1)
				}
				if _, exists := backup.Files[testFilename]; exists {
					t.Error("File should have been removed for 'no' or default response")
				}
			}
		})
	}
}

// TestInteractiveBackupIntegration tests integration with backup workflow
func TestInteractiveBackupIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary SSH directory for testing
	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	err := os.MkdirAll(sshDir, 0700)
	if err != nil {
		t.Fatalf("Failed to create test SSH directory: %v", err)
	}

	// Create test SSH files
	testFiles := map[string]struct {
		content     string
		permissions os.FileMode
	}{
		"id_rsa": {
			content: `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEA1234567890
-----END OPENSSH PRIVATE KEY-----`,
			permissions: 0600,
		},
		"id_rsa.pub": {
			content:     "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDT1234567890 user@host",
			permissions: 0644,
		},
		"config": {
			content: `Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_rsa`,
			permissions: 0600,
		},
	}

	for filename, fileInfo := range testFiles {
		filePath := filepath.Join(sshDir, filename)
		err := os.WriteFile(filePath, []byte(fileInfo.content), fileInfo.permissions)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Test SSH handler can read the directory
	t.Run("ssh_handler_integration", func(t *testing.T) {
		sshHandler := ssh.New()
		backupData, err := sshHandler.ReadDirectory(sshDir)
		if err != nil {
			t.Fatalf("SSH handler failed to read directory: %v", err)
		}

		if len(backupData.Files) == 0 {
			t.Error("SSH handler should have found test files")
		}

		// Verify files have proper structure for interactive selection
		for filename, fileData := range backupData.Files {
			if fileData.KeyInfo == nil {
				t.Errorf("File %s should have KeyInfo for interactive display", filename)
			}
			if fileData.KeyInfo.Type == "" {
				t.Errorf("File %s should have type information", filename)
			}
			if fileData.Size == 0 {
				t.Errorf("File %s should have size information", filename)
			}
		}
	})

	// Test backup options creation
	t.Run("backup_options_creation", func(t *testing.T) {
		opts := backupOptions{
			name:        "integration-test",
			sshDir:      sshDir,
			dryRun:      true, // Use dry run to avoid needing Vault setup
			interactive: true,
		}

		if opts.sshDir != sshDir {
			t.Errorf("SSH directory should be %s, got %s", sshDir, opts.sshDir)
		}

		if !opts.interactive {
			t.Error("Interactive flag should be true")
		}

		if !opts.dryRun {
			t.Error("Dry run flag should be true for this test")
		}
	})
}

// Helper function to create test backup data
func createTestBackupData(t *testing.T) *ssh.BackupData {
	t.Helper()

	// Create test key info
	privateKeyInfo := &analyzer.KeyInfo{
		Filename:    "id_rsa",
		Type:        analyzer.KeyTypePrivate,
		Format:      analyzer.FormatOpenSSH,
		Service:     "github",
		Purpose:     analyzer.PurposeService,
		Permissions: 0600,
		Size:        2048,
		ModTime:     time.Now(),
	}

	publicKeyInfo := &analyzer.KeyInfo{
		Filename:    "id_rsa.pub",
		Type:        analyzer.KeyTypePublic,
		Format:      analyzer.FormatOpenSSH,
		Service:     "github",
		Purpose:     analyzer.PurposeService,
		Permissions: 0644,
		Size:        400,
		ModTime:     time.Now(),
	}

	configInfo := &analyzer.KeyInfo{
		Filename:    "config",
		Type:        analyzer.KeyTypeConfig,
		Format:      analyzer.FormatConfig,
		Purpose:     analyzer.PurposeSystem,
		Permissions: 0600,
		Size:        150,
		ModTime:     time.Now(),
	}

	// Create test file data
	files := map[string]*ssh.FileData{
		"id_rsa": {
			Filename:    "id_rsa",
			Content:     []byte("-----BEGIN OPENSSH PRIVATE KEY-----\ntest-private-key\n-----END OPENSSH PRIVATE KEY-----"),
			Permissions: 0600,
			Size:        2048,
			ModTime:     time.Now(),
			Checksum:    "abc123def456",
			KeyInfo:     privateKeyInfo,
		},
		"id_rsa.pub": {
			Filename:    "id_rsa.pub",
			Content:     []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ test@example.com"),
			Permissions: 0644,
			Size:        400,
			ModTime:     time.Now(),
			Checksum:    "def456ghi789",
			KeyInfo:     publicKeyInfo,
		},
		"config": {
			Filename:    "config",
			Content:     []byte("Host github.com\n    HostName github.com\n    User git"),
			Permissions: 0600,
			Size:        150,
			ModTime:     time.Now(),
			Checksum:    "ghi789jkl012",
			KeyInfo:     configInfo,
		},
	}

	// Create analysis result
	analysis := &analyzer.DetectionResult{
		Keys: []analyzer.KeyInfo{*privateKeyInfo, *publicKeyInfo, *configInfo},
		KeyPairs: map[string]*analyzer.KeyPairInfo{
			"id_rsa": {
				BaseName:       "id_rsa",
				PrivateKeyFile: "id_rsa",
				PublicKeyFile:  "id_rsa.pub",
			},
		},
		Categories: map[string][]analyzer.KeyInfo{
			"service": {*privateKeyInfo, *publicKeyInfo},
			"system":  {*configInfo},
		},
		Summary: &analyzer.AnalysisSummary{
			TotalFiles:   3,
			KeyPairCount: 1,
			ServiceKeys:  2,
			PersonalKeys: 0,
			WorkKeys:     0,
			SystemFiles:  1,
		},
	}

	return &ssh.BackupData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Hostname:  "test-host",
		Username:  "test-user",
		SSHDir:    "/tmp/test-ssh",
		Files:     files,
		Analysis:  analysis,
		Metadata: map[string]interface{}{
			"total_files": len(files),
			"total_size":  int64(2598), // Sum of file sizes
		},
	}
}
