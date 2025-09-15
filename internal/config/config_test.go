package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg == nil {
		t.Fatal("Default() returned nil")
	}

	if cfg.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", cfg.Version)
	}

	if cfg.Vault.Address != "http://localhost:8200" {
		t.Errorf("Expected default vault address, got %s", cfg.Vault.Address)
	}

	if cfg.Vault.MountPath != "ssh-backups" {
		t.Errorf("Expected mount path ssh-backups, got %s", cfg.Vault.MountPath)
	}

	if cfg.Security.Algorithm != "AES-256-GCM" {
		t.Errorf("Expected AES-256-GCM algorithm, got %s", cfg.Security.Algorithm)
	}

	if cfg.Security.Iterations != 100000 {
		t.Errorf("Expected 100000 iterations, got %d", cfg.Security.Iterations)
	}

	if !cfg.Security.PerFileEncrypt {
		t.Error("Expected per file encryption to be enabled")
	}

	if !cfg.Security.VerifyIntegrity {
		t.Error("Expected integrity verification to be enabled")
	}
}

func TestConfig_Save(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "ssh-vault-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := Default()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err = cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Read file content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Check content contains expected values
	contentStr := string(content)
	if !contains(contentStr, "version: \"1.0\"") {
		t.Error("Config file doesn't contain version")
	}

	if !contains(contentStr, "address: http://localhost:8200") {
		t.Error("Config file doesn't contain vault address")
	}

	if !contains(contentStr, "algorithm: AES-256-GCM") {
		t.Error("Config file doesn't contain encryption algorithm")
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()

	if path == "" {
		t.Error("GetConfigPath() returned empty string")
	}

	if !contains(path, ".ssh-vault-keeper") {
		t.Error("Config path doesn't contain .ssh-vault-keeper directory")
	}

	if !contains(path, "config.yaml") {
		t.Error("Config path doesn't end with config.yaml")
	}
}

func TestVaultConfig_Fields(t *testing.T) {
	cfg := Default()

	// Test default values
	if cfg.Vault.Address == "" {
		t.Error("Vault address is empty")
	}

	if cfg.Vault.TokenFile == "" {
		t.Error("Token file path is empty")
	}

	if cfg.Vault.MountPath == "" {
		t.Error("Mount path is empty")
	}

	if cfg.Vault.TLSSkipVerify {
		t.Error("TLS skip verify should be false by default")
	}
}

func TestBackupConfig_Fields(t *testing.T) {
	cfg := Default()

	if cfg.Backup.SSHDir == "" {
		t.Error("SSH directory is empty")
	}

	if !cfg.Backup.HostnamePrefix {
		t.Error("Hostname prefix should be enabled by default")
	}

	if cfg.Backup.RetentionCount <= 0 {
		t.Error("Retention count should be positive")
	}

	if len(cfg.Backup.IncludePatterns) == 0 {
		t.Error("No include patterns defined")
	}

	if len(cfg.Backup.ExcludePatterns) == 0 {
		t.Error("No exclude patterns defined")
	}

	// Check for common include patterns
	includePatterns := cfg.Backup.IncludePatterns
	expectedPatterns := []string{"*.rsa", "*.pub", "config", "known_hosts*"}

	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range includePatterns {
			if pattern == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected include pattern %s not found", expected)
		}
	}
}

func TestSecurityConfig_Fields(t *testing.T) {
	cfg := Default()

	if cfg.Security.Algorithm == "" {
		t.Error("Security algorithm is empty")
	}

	if cfg.Security.KeyDerivation == "" {
		t.Error("Key derivation is empty")
	}

	if cfg.Security.Iterations <= 0 {
		t.Error("Iterations should be positive")
	}

	if cfg.Security.Iterations < 10000 {
		t.Error("Iterations should be at least 10000 for security")
	}
}

func TestLoggingConfig_Fields(t *testing.T) {
	cfg := Default()

	if cfg.Logging.Level == "" {
		t.Error("Logging level is empty")
	}

	if cfg.Logging.Format == "" {
		t.Error("Logging format is empty")
	}

	// Check valid log levels
	validLevels := []string{"debug", "info", "warn", "error"}
	levelValid := false
	for _, level := range validLevels {
		if cfg.Logging.Level == level {
			levelValid = true
			break
		}
	}
	if !levelValid {
		t.Errorf("Invalid log level: %s", cfg.Logging.Level)
	}
}

func TestDetectorConfig_Fields(t *testing.T) {
	cfg := Default()

	if len(cfg.Detectors.Enabled) == 0 {
		t.Error("No detectors enabled")
	}

	// Check for expected detectors
	expectedDetectors := []string{"rsa", "pem", "openssh", "config", "hosts"}
	for _, expected := range expectedDetectors {
		found := false
		for _, detector := range cfg.Detectors.Enabled {
			if detector == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected detector %s not found", expected)
		}
	}

	if len(cfg.Detectors.ServiceMapping) == 0 {
		t.Error("No service mappings defined")
	}

	if len(cfg.Detectors.PurposeRules) == 0 {
		t.Error("No purpose rules defined")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || s[0:len(substr)] == substr || contains(s[1:], substr))
}

// Simple substring check since we're not importing strings package in tests
func simpleContains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
