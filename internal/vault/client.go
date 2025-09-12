package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// Load token
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

// loadToken loads Vault token from file
func loadToken(tokenFile string) (string, error) {
	// Expand home directory
	if strings.HasPrefix(tokenFile, "~/") {
		homeDir, _ := os.UserHomeDir()
		tokenFile = filepath.Join(homeDir, tokenFile[2:])
	}

	token, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", fmt.Errorf("failed to read token file %s: %w", tokenFile, err)
	}

	return strings.TrimSpace(string(token)), nil
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
	path := fmt.Sprintf("%s/data/%s/backups/%s", c.mountPath, c.basePath, backupName)

	_, err := c.client.Logical().Delete(path)
	if err != nil {
		return fmt.Errorf("failed to delete backup %s: %w", backupName, err)
	}

	log.Info().
		Str("backup", backupName).
		Msg("Backup deleted successfully")

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
