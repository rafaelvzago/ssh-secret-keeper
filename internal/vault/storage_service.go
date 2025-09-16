package vault

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-secret-keeper/internal/config"
)

// StorageService provides Vault storage functionality following SRP
type StorageService struct {
	client    *api.Client
	mountPath string
	basePath  string
}

// NewStorageService creates a new Vault storage service
func NewStorageService(cfg *config.VaultConfig) (*StorageService, error) {
	// Create Vault client with proper configuration
	client, err := createVaultClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	// Generate base path for this client
	basePath, err := generateBasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to generate base path: %w", err)
	}

	service := &StorageService{
		client:    client,
		mountPath: cfg.MountPath,
		basePath:  basePath,
	}

	log.Info().
		Str("address", cfg.Address).
		Str("mount", cfg.MountPath).
		Str("base_path", basePath).
		Msg("Vault storage service initialized")

	return service, nil
}

// TestConnection tests the Vault connection and permissions
func (s *StorageService) TestConnection(ctx context.Context) error {
	// Test authentication by getting self info
	_, err := s.client.Auth().Token().LookupSelfWithContext(ctx)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	log.Debug().Msg("Vault connection test successful")
	return nil
}

// EnsureMountExists ensures the KV mount exists
func (s *StorageService) EnsureMountExists(ctx context.Context) error {
	mounts, err := s.client.Sys().ListMountsWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to list mounts: %w", err)
	}

	mountPath := s.mountPath + "/"
	if _, exists := mounts[mountPath]; exists {
		log.Debug().Str("mount", s.mountPath).Msg("Mount already exists")
		return nil
	}

	// Create KV v2 mount
	err = s.client.Sys().MountWithContext(ctx, s.mountPath, &api.MountInput{
		Type:        "kv",
		Description: "SSH Secret Keeper storage",
		Options: map[string]string{
			"version": "2",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create mount %s: %w", s.mountPath, err)
	}

	log.Info().Str("mount", s.mountPath).Msg("Created KV v2 mount")
	return nil
}

// StoreBackup stores encrypted SSH backup data
func (s *StorageService) StoreBackup(ctx context.Context, backupName string, data map[string]interface{}) error {
	path := s.buildBackupPath(backupName)

	// Add timestamp and metadata
	wrappedData := map[string]interface{}{
		"data": data,
		"metadata": map[string]interface{}{
			"stored_at": time.Now().Format(time.RFC3339),
			"path":      path,
		},
	}

	_, err := s.client.Logical().WriteWithContext(ctx, path, wrappedData)
	if err != nil {
		return fmt.Errorf("failed to store backup %s: %w", backupName, err)
	}

	log.Info().
		Str("backup", backupName).
		Str("path", path).
		Msg("Backup stored successfully")

	return nil
}

// GetBackup retrieves encrypted SSH backup data
func (s *StorageService) GetBackup(ctx context.Context, backupName string) (map[string]interface{}, error) {
	path := s.buildBackupPath(backupName)

	secret, err := s.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup %s: %w", backupName, err)
	}

	if secret == nil {
		return nil, fmt.Errorf("backup %s not found", backupName)
	}

	// Extract data from KV v2 format
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid backup data format for %s", backupName)
	}

	log.Debug().
		Str("backup", backupName).
		Msg("Backup retrieved successfully")

	return data, nil
}

// ListBackups lists all available backups
func (s *StorageService) ListBackups(ctx context.Context) ([]string, error) {
	path := s.buildBackupListPath()

	secret, err := s.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	if secret == nil {
		return []string{}, nil
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	var backups []string
	for _, key := range keys {
		if keyStr, ok := key.(string); ok {
			backups = append(backups, keyStr)
		}
	}

	log.Debug().
		Int("count", len(backups)).
		Msg("Listed backups")

	return backups, nil
}

// DeleteBackup deletes a backup
func (s *StorageService) DeleteBackup(ctx context.Context, backupName string) error {
	path := s.buildBackupPath(backupName)

	_, err := s.client.Logical().DeleteWithContext(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete backup %s: %w", backupName, err)
	}

	log.Info().
		Str("backup", backupName).
		Msg("Backup deleted successfully")

	return nil
}

// StoreMetadata stores backup metadata
func (s *StorageService) StoreMetadata(ctx context.Context, metadata map[string]interface{}) error {
	path := s.buildMetadataPath()

	secretData := map[string]interface{}{
		"data":       metadata,
		"updated_at": time.Now().Format(time.RFC3339),
	}

	_, err := s.client.Logical().WriteWithContext(ctx, path, secretData)
	if err != nil {
		return fmt.Errorf("failed to store metadata: %w", err)
	}

	log.Debug().Msg("Metadata stored successfully")
	return nil
}

// GetMetadata retrieves backup metadata
func (s *StorageService) GetMetadata(ctx context.Context) (map[string]interface{}, error) {
	path := s.buildMetadataPath()

	secret, err := s.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	if secret == nil {
		return make(map[string]interface{}), nil
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return make(map[string]interface{}), nil
	}

	return data, nil
}

// GetBasePath returns the base path for this client
func (s *StorageService) GetBasePath() string {
	return s.basePath
}

// Close cleans up the client resources
func (s *StorageService) Close() {
	// Clear token from memory
	if s.client != nil {
		s.client.SetToken("")
	}
	log.Debug().Msg("Vault storage service closed")
}

// Private helper methods

func (s *StorageService) buildBackupPath(backupName string) string {
	return fmt.Sprintf("%s/data/%s/backups/%s", s.mountPath, s.basePath, backupName)
}

func (s *StorageService) buildBackupListPath() string {
	return fmt.Sprintf("%s/metadata/%s/backups", s.mountPath, s.basePath)
}

func (s *StorageService) buildMetadataPath() string {
	return fmt.Sprintf("%s/data/%s/metadata", s.mountPath, s.basePath)
}
