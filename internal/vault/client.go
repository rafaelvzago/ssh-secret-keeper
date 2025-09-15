package vault

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-vault-keeper/internal/config"
)

// Client wraps HashiCorp Vault API client with SSH-specific operations
type Client struct {
	client    *api.Client
	mountPath string
	basePath  string
}

// New creates a new Vault client from configuration
func New(cfg *config.VaultConfig) (*Client, error) {
	// Create Vault client
	vaultConfig := api.DefaultConfig()
	vaultConfig.Address = cfg.Address

	if cfg.TLSSkipVerify {
		vaultConfig.ConfigureTLS(&api.TLSConfig{
			Insecure: true,
		})
	}

	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	// Set namespace if provided
	if cfg.Namespace != "" {
		client.SetNamespace(cfg.Namespace)
	}

	// Load token using client factory function
	token, err := loadToken(cfg.TokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load Vault token: %w", err)
	}
	client.SetToken(token)

	// Test connection
	if err := testConnection(client); err != nil {
		return nil, fmt.Errorf("vault connection test failed: %w", err)
	}

	// Generate base path (hostname + username)
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = "unknown"
	}
	basePath := fmt.Sprintf("users/%s-%s", hostname, username)

	log.Info().
		Str("address", cfg.Address).
		Str("mount", cfg.MountPath).
		Str("base_path", basePath).
		Msg("Vault client initialized")

	return &Client{
		client:    client,
		mountPath: cfg.MountPath,
		basePath:  basePath,
	}, nil
}

// testConnection tests the Vault connection and permissions
func testConnection(client *api.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test authentication by getting self info
	_, err := client.Auth().Token().LookupSelfWithContext(ctx)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	return nil
}

// EnsureMountExists ensures the KV mount exists
func (c *Client) EnsureMountExists() error {
	mounts, err := c.client.Sys().ListMounts()
	if err != nil {
		return fmt.Errorf("failed to list mounts: %w", err)
	}

	mountPath := c.mountPath + "/"
	if _, exists := mounts[mountPath]; exists {
		log.Debug().Str("mount", c.mountPath).Msg("Mount already exists")
		return nil
	}

	// Create KV v2 mount
	err = c.client.Sys().Mount(c.mountPath, &api.MountInput{
		Type:        "kv",
		Description: "SSH Vault Keeper storage",
		Options: map[string]string{
			"version": "2",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create mount %s: %w", c.mountPath, err)
	}

	log.Info().Str("mount", c.mountPath).Msg("Created KV v2 mount")
	return nil
}

// StoreBackup stores encrypted SSH backup data
func (c *Client) StoreBackup(backupName string, data map[string]interface{}) error {
	path := fmt.Sprintf("%s/data/%s/backups/%s", c.mountPath, c.basePath, backupName)

	// Wrap data in KV v2 format
	secretData := map[string]interface{}{
		"data": data,
	}

	_, err := c.client.Logical().Write(path, secretData)
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
func (c *Client) GetBackup(backupName string) (map[string]interface{}, error) {
	path := fmt.Sprintf("%s/data/%s/backups/%s", c.mountPath, c.basePath, backupName)

	secret, err := c.client.Logical().Read(path)
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
func (c *Client) ListBackups() ([]string, error) {
	path := fmt.Sprintf("%s/metadata/%s/backups", c.mountPath, c.basePath)

	secret, err := c.client.Logical().List(path)
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
func (c *Client) DeleteBackup(backupName string) error {
	// Delete from data path
	dataPath := fmt.Sprintf("%s/data/%s/backups/%s", c.mountPath, c.basePath, backupName)
	_, err := c.client.Logical().Delete(dataPath)
	if err != nil {
		return fmt.Errorf("failed to delete backup %s: %w", backupName, err)
	}

	log.Info().
		Str("backup", backupName).
		Msg("Backup deleted successfully")

	return nil
}

// ForceDeleteBackup deletes a backup and waits for metadata to update
func (c *Client) ForceDeleteBackup(backupName string) error {
	// First, delete the backup data
	if err := c.DeleteBackup(backupName); err != nil {
		return err
	}

	// Also try to delete from metadata path (some Vault versions may need this)
	c.DeleteBackupMetadata(backupName)

	// Wait for metadata to update with retries
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		time.Sleep(time.Duration(i+1) * time.Second) // Increasing delay: 1s, 2s, 3s, 4s, 5s

		backups, err := c.ListBackups()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to verify deletion")
			continue
		}

		// Check if backup still exists
		stillExists := false
		for _, backup := range backups {
			if backup == backupName {
				stillExists = true
				break
			}
		}

		if !stillExists {
			log.Debug().
				Str("backup", backupName).
				Int("retry", i+1).
				Msg("Backup successfully removed from metadata")
			return nil
		}

		log.Debug().
			Str("backup", backupName).
			Int("retry", i+1).
			Msg("Backup still appears in metadata, retrying...")
	}

	log.Warn().
		Str("backup", backupName).
		Msg("Backup data deleted but still appears in metadata after retries")

	return nil
}

// DeleteBackupMetadata deletes backup metadata entry (for cleanup)
func (c *Client) DeleteBackupMetadata(backupName string) error {
	// Try to delete from metadata path as well (some Vault versions may need this)
	metadataPath := fmt.Sprintf("%s/metadata/%s/backups/%s", c.mountPath, c.basePath, backupName)
	_, err := c.client.Logical().Delete(metadataPath)
	if err != nil {
		log.Debug().
			Str("backup", backupName).
			Err(err).
			Msg("Failed to delete from metadata path (this may be normal)")
	} else {
		log.Debug().
			Str("backup", backupName).
			Msg("Successfully deleted from metadata path")
	}
	return nil
}

// StoreMetadata stores backup metadata
func (c *Client) StoreMetadata(metadata map[string]interface{}) error {
	path := fmt.Sprintf("%s/data/%s/metadata", c.mountPath, c.basePath)

	secretData := map[string]interface{}{
		"data": metadata,
	}

	_, err := c.client.Logical().Write(path, secretData)
	if err != nil {
		return fmt.Errorf("failed to store metadata: %w", err)
	}

	log.Debug().Msg("Metadata stored successfully")
	return nil
}

// GetMetadata retrieves backup metadata
func (c *Client) GetMetadata() (map[string]interface{}, error) {
	path := fmt.Sprintf("%s/data/%s/metadata", c.mountPath, c.basePath)

	secret, err := c.client.Logical().Read(path)
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

// TestConnection tests the connection to Vault
func (c *Client) TestConnection() error {
	return testConnection(c.client)
}

// GetBasePath returns the base path for this client
func (c *Client) GetBasePath() string {
	return c.basePath
}

// Close cleans up the client resources
func (c *Client) Close() {
	// Clear token from memory
	c.client.SetToken("")
	log.Debug().Msg("Vault client closed")
}
