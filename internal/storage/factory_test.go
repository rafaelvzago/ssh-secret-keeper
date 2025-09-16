package storage

import (
	"os"
	"testing"

	"github.com/rzago/ssh-secret-keeper/internal/config"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	if factory == nil {
		t.Fatal("NewFactory() returned nil")
	}
}

func TestFactory_CreateStorage(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name          string
		cfg           *config.Config
		envVaultAddr  string
		envVaultToken string
		wantError     bool
		wantProvider  string
		errorContains string
	}{
		{
			name: "vault provider with config address",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					Provider: "vault",
				},
				Vault: config.VaultConfig{
					Address:   "http://localhost:8200",
					MountPath: "ssh-backups",
					TokenFile: "/dev/null", // Will fail but we're testing factory logic
				},
			},
			envVaultToken: "test-token",
			wantError:     true, // Will fail on actual vault client creation
			wantProvider:  "vault",
		},
		{
			name: "vault provider with env address override",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					Provider: "vault",
				},
				Vault: config.VaultConfig{
					Address:   "http://config-address:8200",
					MountPath: "ssh-backups",
					TokenFile: "/dev/null",
				},
			},
			envVaultAddr:  "http://env-address:8200",
			envVaultToken: "test-token",
			wantError:     true, // Will fail on actual vault client creation
			wantProvider:  "vault",
		},
		{
			name: "default to vault when provider not specified",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					// Provider not specified
				},
				Vault: config.VaultConfig{
					Address:   "http://localhost:8200",
					MountPath: "ssh-backups",
					TokenFile: "/dev/null",
				},
			},
			envVaultToken: "test-token",
			wantError:     true, // Will fail on actual vault client creation
			wantProvider:  "vault",
		},
		{
			name: "vault provider without address",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					Provider: "vault",
				},
				Vault: config.VaultConfig{
					// Address not specified
					MountPath: "ssh-backups",
					TokenFile: "/dev/null",
				},
			},
			wantError:     true,
			errorContains: "vault address not configured",
		},
		{
			name: "onepassword provider not implemented",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					Provider: "onepassword",
				},
			},
			wantError:     true,
			errorContains: "1Password provider not implemented yet",
		},
		{
			name: "s3 provider not implemented",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					Provider: "s3",
				},
			},
			wantError:     true,
			errorContains: "S3 provider not implemented yet",
		},
		{
			name: "unsupported provider",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					Provider: "unsupported",
				},
			},
			wantError:     true,
			errorContains: "unsupported storage provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalAddr := os.Getenv("VAULT_ADDR")
			originalToken := os.Getenv("VAULT_TOKEN")
			defer func() {
				os.Setenv("VAULT_ADDR", originalAddr)
				os.Setenv("VAULT_TOKEN", originalToken)
			}()

			// Set test environment
			if tt.envVaultAddr != "" {
				os.Setenv("VAULT_ADDR", tt.envVaultAddr)
			} else {
				os.Unsetenv("VAULT_ADDR")
			}

			if tt.envVaultToken != "" {
				os.Setenv("VAULT_TOKEN", tt.envVaultToken)
			} else {
				os.Unsetenv("VAULT_TOKEN")
			}

			provider, err := factory.CreateStorage(tt.cfg)

			if tt.wantError {
				if err == nil {
					t.Errorf("CreateStorage() expected error but got none")
					return
				}

				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("CreateStorage() error = %v, should contain %s", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("CreateStorage() unexpected error: %v", err)
				return
			}

			if provider == nil {
				t.Error("CreateStorage() returned nil provider")
				return
			}

			// Test provider type if specified
			if tt.wantProvider != "" {
				if provider.GetProviderType() != tt.wantProvider {
					t.Errorf("CreateStorage() provider type = %s, want %s",
						provider.GetProviderType(), tt.wantProvider)
				}
			}

			// Clean up
			provider.Close()
		})
	}
}


func TestFactory_CreateStorage_DefaultProvider(t *testing.T) {
	factory := NewFactory()

	cfg := &config.Config{
		Storage: config.StorageConfig{
			// Provider not set - should default to vault
		},
		Vault: config.VaultConfig{
			Address:   "http://localhost:8200",
			MountPath: "ssh-backups",
			TokenFile: "/dev/null",
		},
	}

	// Set token for the test
	os.Setenv("VAULT_TOKEN", "test-token")
	defer os.Unsetenv("VAULT_TOKEN")

	// Should default to vault provider
	_, err := factory.CreateStorage(cfg)

	// Will fail on vault client creation, but should have set provider to vault
	if cfg.Storage.Provider != "vault" {
		t.Errorf("Provider not defaulted to vault: got %s", cfg.Storage.Provider)
	}

	// Should get vault-related error, not unsupported provider error
	if err != nil && contains(err.Error(), "unsupported storage provider") {
		t.Errorf("Got unsupported provider error when should default to vault: %v", err)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
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
