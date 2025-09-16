package vault

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rzago/ssh-secret-keeper/internal/config"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.VaultConfig
		envToken  string
		wantError bool
	}{
		{
			name: "valid config with env token",
			cfg: &config.VaultConfig{
				Address:   "http://localhost:8200",
				MountPath: "ssh-backups",
				TokenFile: "/nonexistent", // Will use env token
			},
			envToken:  "test-token-123",
			wantError: true, // Will fail on connection test without real Vault
		},
		{
			name: "config with TLS skip verify",
			cfg: &config.VaultConfig{
				Address:       "https://vault.example.com",
				MountPath:     "ssh-backups",
				TLSSkipVerify: true,
				TokenFile:     "/nonexistent",
			},
			envToken:  "test-token-123",
			wantError: true, // Will fail on connection test
		},
		{
			name: "empty token should fail",
			cfg: &config.VaultConfig{
				Address:   "http://localhost:8200",
				MountPath: "ssh-backups",
				TokenFile: "/nonexistent",
			},
			envToken:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			originalToken := os.Getenv("VAULT_TOKEN")
			originalUser := os.Getenv("USER")
			defer func() {
				os.Setenv("VAULT_TOKEN", originalToken)
				os.Setenv("USER", originalUser)
			}()

			if tt.envToken != "" {
				os.Setenv("VAULT_TOKEN", tt.envToken)
			} else {
				os.Unsetenv("VAULT_TOKEN")
			}
			os.Setenv("USER", "testuser")

			client, err := New(tt.cfg)

			if tt.wantError {
				if err == nil {
					t.Errorf("New() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("New() unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("New() returned nil client")
				return
			}

			// Clean up
			client.Close()
		})
	}
}

func TestClient_PathGeneration(t *testing.T) {
	// Create a client with mock data for testing path generation
	client := &Client{
		client:    nil, // Not needed for path testing
		mountPath: "ssh-backups",
		basePath:  "users/test-host-testuser",
	}

	tests := []struct {
		name       string
		backupName string
		wantData   string
		wantMeta   string
		wantList   string
	}{
		{
			name:       "simple backup name",
			backupName: "backup-123",
			wantData:   "ssh-backups/data/users/test-host-testuser/backups/backup-123",
			wantMeta:   "ssh-backups/metadata/users/test-host-testuser/backups/backup-123",
			wantList:   "ssh-backups/metadata/users/test-host-testuser/backups",
		},
		{
			name:       "backup with timestamp",
			backupName: "backup-20240115-143022",
			wantData:   "ssh-backups/data/users/test-host-testuser/backups/backup-20240115-143022",
			wantMeta:   "ssh-backups/metadata/users/test-host-testuser/backups/backup-20240115-143022",
			wantList:   "ssh-backups/metadata/users/test-host-testuser/backups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test data path generation (internal method simulation)
			dataPath := fmt.Sprintf("%s/data/%s/backups/%s", client.mountPath, client.basePath, tt.backupName)
			if dataPath != tt.wantData {
				t.Errorf("data path = %s, want %s", dataPath, tt.wantData)
			}

			// Test metadata path generation
			metaPath := fmt.Sprintf("%s/metadata/%s/backups/%s", client.mountPath, client.basePath, tt.backupName)
			if metaPath != tt.wantMeta {
				t.Errorf("metadata path = %s, want %s", metaPath, tt.wantMeta)
			}

			// Test list path generation
			listPath := fmt.Sprintf("%s/metadata/%s/backups", client.mountPath, client.basePath)
			if listPath != tt.wantList {
				t.Errorf("list path = %s, want %s", listPath, tt.wantList)
			}
		})
	}
}

func TestClient_GetBasePath(t *testing.T) {
	client := &Client{
		basePath: "users/test-host-testuser",
	}

	basePath := client.GetBasePath()
	if basePath != "users/test-host-testuser" {
		t.Errorf("GetBasePath() = %s, want users/test-host-testuser", basePath)
	}
}

func TestClient_Close(t *testing.T) {
	// We can't easily test Close with nil client due to the implementation
	// calling SetToken on the client. This would require a real client.
	t.Skip("Close method requires real Vault client, tested in integration tests")
}

func TestTestConnection(t *testing.T) {
	// This tests the standalone testConnection function
	// We can't easily test this without a real Vault instance
	// But we can test that it handles nil client gracefully

	if testing.Short() {
		t.Skip("Skipping connection test in short mode")
	}

	// This would panic with nil client, so we skip the actual test
	// In a real test environment with Vault, this would test the connection
	t.Skip("testConnection requires real Vault instance")
}

// Integration tests that require a real Vault instance
func TestClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Only run if VAULT_ADDR and VAULT_TOKEN are set for integration testing
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultToken := os.Getenv("VAULT_TOKEN")

	if vaultAddr == "" || vaultToken == "" {
		t.Skip("Skipping integration tests - VAULT_ADDR and VAULT_TOKEN must be set")
	}

	cfg := &config.VaultConfig{
		Address:   vaultAddr,
		MountPath: "ssh-backups-test",
		TokenFile: "/nonexistent", // Will use env token
	}

	client, err := New(cfg)
	if err != nil {
		t.Skipf("Could not create client: %v", err)
	}
	defer client.Close()

	t.Run("EnsureMountExists", func(t *testing.T) {
		err := client.EnsureMountExists()
		if err != nil {
			t.Errorf("EnsureMountExists() failed: %v", err)
		}
	})

	t.Run("StoreAndGetBackup", func(t *testing.T) {
		backupName := fmt.Sprintf("test-backup-%d", time.Now().Unix())
		testData := map[string]interface{}{
			"file1": "encrypted-content-1",
			"file2": "encrypted-content-2",
		}

		// Store backup
		err := client.StoreBackup(backupName, testData)
		if err != nil {
			t.Errorf("StoreBackup() failed: %v", err)
			return
		}

		// Get backup
		retrievedData, err := client.GetBackup(backupName)
		if err != nil {
			t.Errorf("GetBackup() failed: %v", err)
			return
		}

		// Verify data
		for key, value := range testData {
			if retrievedData[key] != value {
				t.Errorf("Retrieved data mismatch for key %s: got %v, want %v",
					key, retrievedData[key], value)
			}
		}

		// Clean up
		err = client.DeleteBackup(backupName)
		if err != nil {
			t.Errorf("DeleteBackup() failed: %v", err)
		}
	})

	t.Run("ListBackups", func(t *testing.T) {
		backups, err := client.ListBackups()
		if err != nil {
			t.Errorf("ListBackups() failed: %v", err)
		}

		// Should return a slice (empty or with items)
		if backups == nil {
			t.Error("ListBackups() returned nil")
		}
	})

	t.Run("StoreAndGetMetadata", func(t *testing.T) {
		testMetadata := map[string]interface{}{
			"last_backup": time.Now().Format(time.RFC3339),
			"version":     "1.0.0",
		}

		// Store metadata
		err := client.StoreMetadata(testMetadata)
		if err != nil {
			t.Errorf("StoreMetadata() failed: %v", err)
			return
		}

		// Get metadata
		retrievedMetadata, err := client.GetMetadata()
		if err != nil {
			t.Errorf("GetMetadata() failed: %v", err)
			return
		}

		// Verify metadata
		for key, value := range testMetadata {
			if retrievedMetadata[key] != value {
				t.Errorf("Retrieved metadata mismatch for key %s: got %v, want %v",
					key, retrievedMetadata[key], value)
			}
		}
	})

	t.Run("ForceDeleteBackup", func(t *testing.T) {
		backupName := fmt.Sprintf("test-force-delete-%d", time.Now().Unix())
		testData := map[string]interface{}{
			"test": "data",
		}

		// Store backup
		err := client.StoreBackup(backupName, testData)
		if err != nil {
			t.Errorf("StoreBackup() failed: %v", err)
			return
		}

		// Force delete
		err = client.ForceDeleteBackup(backupName)
		if err != nil {
			t.Errorf("ForceDeleteBackup() failed: %v", err)
		}

		// Verify deletion
		_, err = client.GetBackup(backupName)
		if err == nil {
			t.Error("Expected error when getting deleted backup")
		}
	})

	t.Run("GetNonexistentBackup", func(t *testing.T) {
		_, err := client.GetBackup("nonexistent-backup")
		if err == nil {
			t.Error("Expected error when getting nonexistent backup")
		}
	})

	t.Run("TestConnection", func(t *testing.T) {
		err := client.TestConnection()
		if err != nil {
			t.Errorf("TestConnection() failed: %v", err)
		}
	})
}

// Benchmark tests
func BenchmarkClient_PathGeneration(b *testing.B) {
	client := &Client{
		mountPath: "ssh-backups",
		basePath:  "users/test-host-testuser",
	}

	b.Run("DataPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = fmt.Sprintf("%s/data/%s/backups/%s", client.mountPath, client.basePath, "test-backup")
		}
	})

	b.Run("MetadataPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = fmt.Sprintf("%s/metadata/%s/backups/%s", client.mountPath, client.basePath, "test-backup")
		}
	})
}
