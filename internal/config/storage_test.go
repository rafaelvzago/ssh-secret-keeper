package config

import (
	"testing"
)

func TestStorageConfig_DefaultProvider(t *testing.T) {
	cfg := Default()

	if cfg.Storage.Provider != "vault" {
		t.Errorf("Expected default storage provider 'vault', got '%s'", cfg.Storage.Provider)
	}
}

func TestStorageConfig_VaultConfiguration(t *testing.T) {
	cfg := Default()

	if cfg.Storage.Vault == nil {
		t.Fatal("Expected vault configuration to be present")
	}

	// Test vault config fields
	if cfg.Storage.Vault.Address == "" {
		t.Error("Vault address should not be empty")
	}

	if cfg.Storage.Vault.TokenFile == "" {
		t.Error("Vault token file should not be empty")
	}

	if cfg.Storage.Vault.MountPath == "" {
		t.Error("Vault mount path should not be empty")
	}

	// Test default values
	if cfg.Storage.Vault.MountPath != "ssh-backups" {
		t.Errorf("Expected default mount path 'ssh-backups', got '%s'", cfg.Storage.Vault.MountPath)
	}

	if cfg.Storage.Vault.TLSSkipVerify {
		t.Error("Expected TLSSkipVerify to default to false")
	}
}

func TestStorageConfig_OnePasswordConfiguration(t *testing.T) {
	// Test OnePassword config structure
	opConfig := &OnePasswordConfig{
		ServerURL:    "https://1password.example.com",
		Token:        "test-token",
		VaultID:      "vault-123",
		ItemTemplate: "SSH Key",
	}

	storageConfig := StorageConfig{
		Provider:    "onepassword",
		OnePassword: opConfig,
	}

	if storageConfig.Provider != "onepassword" {
		t.Error("OnePassword provider not set correctly")
	}

	if storageConfig.OnePassword == nil {
		t.Fatal("OnePassword config should not be nil")
	}

	if storageConfig.OnePassword.ServerURL != "https://1password.example.com" {
		t.Error("OnePassword server URL not set correctly")
	}

	if storageConfig.OnePassword.Token != "test-token" {
		t.Error("OnePassword token not set correctly")
	}

	if storageConfig.OnePassword.VaultID != "vault-123" {
		t.Error("OnePassword vault ID not set correctly")
	}
}

func TestStorageConfig_S3Configuration(t *testing.T) {
	// Test S3 config structure
	s3Config := &S3Config{
		Endpoint:        "https://s3.amazonaws.com",
		Region:          "us-east-1",
		Bucket:          "ssh-backups",
		AccessKeyID:     "AKIA...",
		SecretAccessKey: "secret...",
		Prefix:          "backups/",
	}

	storageConfig := StorageConfig{
		Provider: "s3",
		S3:       s3Config,
	}

	if storageConfig.Provider != "s3" {
		t.Error("S3 provider not set correctly")
	}

	if storageConfig.S3 == nil {
		t.Fatal("S3 config should not be nil")
	}

	if storageConfig.S3.Endpoint != "https://s3.amazonaws.com" {
		t.Error("S3 endpoint not set correctly")
	}

	if storageConfig.S3.Region != "us-east-1" {
		t.Error("S3 region not set correctly")
	}

	if storageConfig.S3.Bucket != "ssh-backups" {
		t.Error("S3 bucket not set correctly")
	}
}

func TestStorageVaultConfig_Fields(t *testing.T) {
	vaultConfig := &VaultConfig{
		Address:       "https://vault.example.com:8200",
		TokenFile:     "/path/to/token",
		MountPath:     "custom-mount",
		Namespace:     "test-namespace",
		TLSSkipVerify: true,
	}

	if vaultConfig.Address != "https://vault.example.com:8200" {
		t.Error("Vault address not set correctly")
	}

	if vaultConfig.TokenFile != "/path/to/token" {
		t.Error("Vault token file not set correctly")
	}

	if vaultConfig.MountPath != "custom-mount" {
		t.Error("Vault mount path not set correctly")
	}

	if vaultConfig.Namespace != "test-namespace" {
		t.Error("Vault namespace not set correctly")
	}

	if !vaultConfig.TLSSkipVerify {
		t.Error("Vault TLSSkipVerify not set correctly")
	}
}

func TestOnePasswordConfig_Fields(t *testing.T) {
	opConfig := &OnePasswordConfig{
		ServerURL:    "https://connect.1password.com",
		Token:        "connect-token-123",
		VaultID:      "vault-abc-123",
		ItemTemplate: "SSH Private Key",
	}

	if opConfig.ServerURL != "https://connect.1password.com" {
		t.Error("OnePassword server URL not set correctly")
	}

	if opConfig.Token != "connect-token-123" {
		t.Error("OnePassword token not set correctly")
	}

	if opConfig.VaultID != "vault-abc-123" {
		t.Error("OnePassword vault ID not set correctly")
	}

	if opConfig.ItemTemplate != "SSH Private Key" {
		t.Error("OnePassword item template not set correctly")
	}
}

func TestS3Config_Fields(t *testing.T) {
	s3Config := &S3Config{
		Endpoint:        "https://minio.example.com",
		Region:          "us-west-2",
		Bucket:          "my-ssh-backups",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin123",
		Prefix:          "ssh-keys/",
	}

	if s3Config.Endpoint != "https://minio.example.com" {
		t.Error("S3 endpoint not set correctly")
	}

	if s3Config.Region != "us-west-2" {
		t.Error("S3 region not set correctly")
	}

	if s3Config.Bucket != "my-ssh-backups" {
		t.Error("S3 bucket not set correctly")
	}

	if s3Config.AccessKeyID != "minioadmin" {
		t.Error("S3 access key ID not set correctly")
	}

	if s3Config.SecretAccessKey != "minioadmin123" {
		t.Error("S3 secret access key not set correctly")
	}

	if s3Config.Prefix != "ssh-keys/" {
		t.Error("S3 prefix not set correctly")
	}
}

func TestStorageConfig_MultipleProviders(t *testing.T) {
	// Test that a storage config can have multiple provider configs
	// but only one should be active based on the Provider field
	storageConfig := StorageConfig{
		Provider: "vault",
		Vault: &VaultConfig{
			Address:   "https://vault.example.com:8200",
			TokenFile: "/vault/token",
			MountPath: "ssh-backups",
		},
		OnePassword: &OnePasswordConfig{
			ServerURL: "https://1password.example.com",
			Token:     "op-token",
			VaultID:   "vault-123",
		},
		S3: &S3Config{
			Endpoint: "https://s3.amazonaws.com",
			Region:   "us-east-1",
			Bucket:   "backup-bucket",
		},
	}

	// Only vault should be active
	if storageConfig.Provider != "vault" {
		t.Error("Expected vault provider to be active")
	}

	// All configs should be present but only vault should be used
	if storageConfig.Vault == nil {
		t.Error("Vault config should be present")
	}

	if storageConfig.OnePassword == nil {
		t.Error("OnePassword config should be present (but not active)")
	}

	if storageConfig.S3 == nil {
		t.Error("S3 config should be present (but not active)")
	}
}

func TestStorageConfig_EmptyProvider(t *testing.T) {
	storageConfig := StorageConfig{
		Provider: "",
		Vault:    nil,
	}

	if storageConfig.Provider != "" {
		t.Error("Provider should be empty")
	}

	if storageConfig.Vault != nil {
		t.Error("Vault config should be nil when not configured")
	}
}
