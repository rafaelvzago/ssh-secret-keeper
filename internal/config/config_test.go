package config

import (
	"os"
	"path/filepath"
	"strings"
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
	tmpDir, err := os.MkdirTemp("", "ssh-secret-config-test")
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
	if !strings.Contains(contentStr, "version: \"1.0\"") {
		t.Error("Config file doesn't contain version")
	}

	if !strings.Contains(contentStr, "address: http://localhost:8200") {
		t.Error("Config file doesn't contain vault address")
	}

	if !strings.Contains(contentStr, "algorithm: AES-256-GCM") {
		t.Error("Config file doesn't contain encryption algorithm")
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()

	if path == "" {
		t.Error("GetConfigPath() returned empty string")
	}

	if !strings.Contains(path, ".ssh-secret-keeper") {
		t.Error("Config path doesn't contain .ssh-secret-keeper directory")
	}

	if !strings.Contains(path, "config.yaml") {
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

func TestConfig_EnvironmentVariableOverrides(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected func(*Config) bool
	}{
		{
			name: "VAULT_ADDR override",
			envVars: map[string]string{
				"VAULT_ADDR": "https://vault.example.com:8200",
			},
			expected: func(cfg *Config) bool {
				return cfg.Vault.Address == "https://vault.example.com:8200"
			},
		},
		{
			name: "SSH_SECRET_VAULT_TOKEN_FILE override",
			envVars: map[string]string{
				"VAULT_ADDR":                  "http://localhost:8200",
				"SSH_SECRET_VAULT_TOKEN_FILE": "/custom/token/path",
			},
			expected: func(cfg *Config) bool {
				return cfg.Vault.TokenFile == "/custom/token/path"
			},
		},
		{
			name: "Multiple environment variables",
			envVars: map[string]string{
				"VAULT_ADDR":                  "https://prod-vault.company.com:8200",
				"SSH_SECRET_VAULT_TOKEN_FILE": "/etc/vault/token",
			},
			expected: func(cfg *Config) bool {
				return cfg.Vault.Address == "https://prod-vault.company.com:8200" &&
					cfg.Vault.TokenFile == "/etc/vault/token"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if !tt.expected(cfg) {
				t.Error("Environment variable override did not work as expected")
			}
		})
	}
}

func TestConfig_RequiredEnvironmentVariables(t *testing.T) {
	// Test that VAULT_ADDR is required
	t.Setenv("VAULT_ADDR", "") // Unset VAULT_ADDR

	_, err := Load()
	if err == nil {
		t.Error("Expected error when VAULT_ADDR is not set")
	}

	if !strings.Contains(err.Error(), "VAULT_ADDR environment variable is required") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestConfig_ConfigFileLoading(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := tempDir + "/config.yaml"

	configContent := `version: "1.0"
storage:
  provider: "vault"
  vault:
    address: "http://test-vault:8200"
    mount_path: "test-mount"
vault:
  address: "http://test-vault:8200"
  mount_path: "test-mount"
backup:
  ssh_dir: "/test/ssh"
  hostname_prefix: false
security:
  algorithm: "AES-256-GCM"
  iterations: 200000
logging:
  level: "debug"
  format: "json"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Set environment variables for the test
	t.Setenv("VAULT_ADDR", "http://localhost:8200")

	// Change to the temp directory so the config file is found
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify config was loaded from file
	if cfg.Security.Iterations != 200000 {
		t.Errorf("Expected iterations 200000 from config file, got %d", cfg.Security.Iterations)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug' from config file, got '%s'", cfg.Logging.Level)
	}

	if cfg.Logging.Format != "json" {
		t.Errorf("Expected log format 'json' from config file, got '%s'", cfg.Logging.Format)
	}
}

func TestConfig_StorageConfigValidation(t *testing.T) {
	cfg := Default()

	// Test default storage config
	if cfg.Storage.Provider != "vault" {
		t.Errorf("Expected default storage provider 'vault', got '%s'", cfg.Storage.Provider)
	}

	if cfg.Storage.Vault == nil {
		t.Error("Expected vault config to be present in storage config")
	}

	// Test that storage vault config matches main vault config
	if cfg.Storage.Vault.Address != cfg.Vault.Address {
		t.Error("Storage vault address should match main vault address")
	}

	if cfg.Storage.Vault.MountPath != cfg.Vault.MountPath {
		t.Error("Storage vault mount path should match main vault mount path")
	}
}
