package vault

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/rzago/ssh-secret-keeper/internal/config"
)

// MockVaultClient provides a mock implementation for testing
type MockVaultClient struct {
	token       string
	address     string
	namespace   string
	authError   error
	writeError  error
	readError   error
	listError   error
	deleteError error
	mountError  error
	storage     map[string]interface{}
	mounts      map[string]*api.MountOutput
}

func NewMockVaultClient() *MockVaultClient {
	return &MockVaultClient{
		storage: make(map[string]interface{}),
		mounts:  make(map[string]*api.MountOutput),
	}
}

func (m *MockVaultClient) SetToken(token string) {
	m.token = token
}

func (m *MockVaultClient) Token() string {
	return m.token
}

func (m *MockVaultClient) Address() string {
	return m.address
}

func (m *MockVaultClient) SetNamespace(namespace string) {
	m.namespace = namespace
}

func (m *MockVaultClient) Headers() map[string][]string {
	headers := make(map[string][]string)
	if m.namespace != "" {
		headers["X-Vault-Namespace"] = []string{m.namespace}
	}
	return headers
}

// Mock auth interface
type MockAuth struct {
	client *MockVaultClient
}

func (m *MockVaultClient) Auth() *MockAuth {
	return &MockAuth{client: m}
}

type MockToken struct {
	client *MockVaultClient
}

func (a *MockAuth) Token() *MockToken {
	return &MockToken{client: a.client}
}

func (t *MockToken) LookupSelfWithContext(ctx context.Context) (*api.Secret, error) {
	if t.client.authError != nil {
		return nil, t.client.authError
	}
	return &api.Secret{}, nil
}

// Mock logical interface
type MockLogical struct {
	client *MockVaultClient
}

func (m *MockVaultClient) Logical() *MockLogical {
	return &MockLogical{client: m}
}

func (l *MockLogical) WriteWithContext(ctx context.Context, path string, data map[string]interface{}) (*api.Secret, error) {
	if l.client.writeError != nil {
		return nil, l.client.writeError
	}
	l.client.storage[path] = data
	return &api.Secret{}, nil
}

func (l *MockLogical) ReadWithContext(ctx context.Context, path string) (*api.Secret, error) {
	if l.client.readError != nil {
		return nil, l.client.readError
	}

	data, exists := l.client.storage[path]
	if !exists {
		return nil, nil
	}

	return &api.Secret{
		Data: data.(map[string]interface{}),
	}, nil
}

func (l *MockLogical) ListWithContext(ctx context.Context, path string) (*api.Secret, error) {
	if l.client.listError != nil {
		return nil, l.client.listError
	}

	var keys []interface{}
	for storedPath := range l.client.storage {
		if len(storedPath) > len(path) && storedPath[:len(path)] == path {
			// Extract the key part after the path
			key := storedPath[len(path):]
			if key[0] == '/' {
				key = key[1:]
			}
			// Only add direct children, not nested paths
			if !contains(key, "/") {
				keys = append(keys, key)
			}
		}
	}

	if len(keys) == 0 {
		return nil, nil
	}

	return &api.Secret{
		Data: map[string]interface{}{
			"keys": keys,
		},
	}, nil
}

func (l *MockLogical) DeleteWithContext(ctx context.Context, path string) (*api.Secret, error) {
	if l.client.deleteError != nil {
		return nil, l.client.deleteError
	}
	delete(l.client.storage, path)
	return &api.Secret{}, nil
}

// Mock sys interface
type MockSys struct {
	client *MockVaultClient
}

func (m *MockVaultClient) Sys() *MockSys {
	return &MockSys{client: m}
}

func (s *MockSys) ListMountsWithContext(ctx context.Context) (map[string]*api.MountOutput, error) {
	if s.client.mountError != nil {
		return nil, s.client.mountError
	}
	return s.client.mounts, nil
}

func (s *MockSys) MountWithContext(ctx context.Context, path string, mountInfo *api.MountInput) error {
	if s.client.mountError != nil {
		return s.client.mountError
	}
	s.client.mounts[path+"/"] = &api.MountOutput{
		Type:        mountInfo.Type,
		Description: mountInfo.Description,
		Options:     mountInfo.Options,
	}
	return nil
}

func TestNewStorageService(t *testing.T) {
	// Set up environment for testing
	os.Setenv("VAULT_TOKEN", "test-token")
	defer os.Unsetenv("VAULT_TOKEN")

	cfg := &config.VaultConfig{
		Address:   "http://localhost:8200",
		MountPath: "ssh-backups",
		TokenFile: "/nonexistent", // Will use env token
	}

	// This will fail because we can't actually connect to Vault in tests
	// But we can test the configuration parsing
	service, err := NewStorageService(cfg)

	// The service creation might succeed but connection test should fail
	if err != nil {
		// Expected - connection should fail without real Vault
		return
	}

	if service != nil {
		defer service.Close()
		// Test connection should fail without real Vault
		err = service.TestConnection(context.Background())
		if err == nil {
			t.Error("Expected connection test to fail without real Vault")
		}
	}

	// Test with empty token
	os.Unsetenv("VAULT_TOKEN")
	_, err = NewStorageService(cfg)
	if err == nil {
		t.Error("Expected error when creating storage service without token")
	}
}

// Create a testable storage service with mock client
func createTestStorageService() *StorageService {
	return &StorageService{
		client:    nil, // We'll test path methods without actual client
		mountPath: "ssh-backups",
		basePath:  "users/test-host-testuser",
	}
}

// Since we can't easily mock the Vault client due to its concrete type,
// let's test the individual methods with a more focused approach

func TestStorageService_Methods(t *testing.T) {
	// Test path building methods
	service := &StorageService{
		mountPath: "ssh-backups",
		basePath:  "users/test-host-testuser",
	}

	t.Run("buildBackupPath", func(t *testing.T) {
		path := service.buildBackupPath("backup-123")
		expected := "ssh-backups/data/users/test-host-testuser/backups/backup-123"
		if path != expected {
			t.Errorf("buildBackupPath() = %s, want %s", path, expected)
		}
	})

	t.Run("buildBackupListPath", func(t *testing.T) {
		path := service.buildBackupListPath()
		expected := "ssh-backups/metadata/users/test-host-testuser/backups"
		if path != expected {
			t.Errorf("buildBackupListPath() = %s, want %s", path, expected)
		}
	})

	t.Run("buildMetadataPath", func(t *testing.T) {
		path := service.buildMetadataPath()
		expected := "ssh-backups/data/users/test-host-testuser/metadata"
		if path != expected {
			t.Errorf("buildMetadataPath() = %s, want %s", path, expected)
		}
	})

	t.Run("GetBasePath", func(t *testing.T) {
		basePath := service.GetBasePath()
		expected := "users/test-host-testuser"
		if basePath != expected {
			t.Errorf("GetBasePath() = %s, want %s", basePath, expected)
		}
	})
}

func TestStorageService_Close(t *testing.T) {
	service := &StorageService{
		client:    nil, // Safe to call Close with nil client
		mountPath: "ssh-backups",
		basePath:  "users/test-host-testuser",
	}

	// Should not panic
	service.Close()
}

// Integration-style tests that would work with a real Vault instance
func TestStorageService_Integration(t *testing.T) {
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

	service, err := NewStorageService(cfg)
	if err != nil {
		t.Skipf("Could not create storage service: %v", err)
	}
	defer service.Close()

	ctx := context.Background()

	t.Run("TestConnection", func(t *testing.T) {
		err := service.TestConnection(ctx)
		if err != nil {
			t.Errorf("TestConnection() failed: %v", err)
		}
	})

	t.Run("EnsureMountExists", func(t *testing.T) {
		err := service.EnsureMountExists(ctx)
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
		err := service.StoreBackup(ctx, backupName, testData)
		if err != nil {
			t.Errorf("StoreBackup() failed: %v", err)
			return
		}

		// Get backup
		retrievedData, err := service.GetBackup(ctx, backupName)
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
		err = service.DeleteBackup(ctx, backupName)
		if err != nil {
			t.Errorf("DeleteBackup() failed: %v", err)
		}
	})

	t.Run("ListBackups", func(t *testing.T) {
		backups, err := service.ListBackups(ctx)
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
		err := service.StoreMetadata(ctx, testMetadata)
		if err != nil {
			t.Errorf("StoreMetadata() failed: %v", err)
			return
		}

		// Get metadata
		retrievedMetadata, err := service.GetMetadata(ctx)
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

	t.Run("GetNonexistentBackup", func(t *testing.T) {
		_, err := service.GetBackup(ctx, "nonexistent-backup")
		if err == nil {
			t.Error("Expected error when getting nonexistent backup")
		}
	})
}

// Benchmark tests
func BenchmarkStorageService_PathBuilding(b *testing.B) {
	service := &StorageService{
		mountPath: "ssh-backups",
		basePath:  "users/test-host-testuser",
	}

	b.Run("buildBackupPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = service.buildBackupPath("test-backup")
		}
	})

	b.Run("buildMetadataPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = service.buildMetadataPath()
		}
	})
}
