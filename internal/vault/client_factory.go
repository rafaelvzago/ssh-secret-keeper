package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/rzago/ssh-secret-keeper/internal/config"
)

// createVaultClient creates and configures a Vault API client
func createVaultClient(cfg *config.VaultConfig) (*api.Client, error) {
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

	// Load and set token
	token, err := loadToken(cfg.TokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load Vault token: %w", err)
	}

	// Ensure we actually have a token
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("empty Vault token: authentication requires a valid token")
	}

	client.SetToken(token)

	return client, nil
}

// loadToken loads Vault token with priority: environment variable > file
func loadToken(tokenFile string) (string, error) {
	// First, try to get token from environment variable
	if envToken := os.Getenv("VAULT_TOKEN"); envToken != "" {
		tokenStr := strings.TrimSpace(envToken)
		if len(tokenStr) > 0 {
			return tokenStr, nil
		}
	}

	// If no environment token, try to read from file
	// Expand home directory
	if strings.HasPrefix(tokenFile, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot resolve home directory: %w", err)
		}
		tokenFile = filepath.Join(homeDir, tokenFile[2:])
	}

	// Clean the path
	tokenFile = filepath.Clean(tokenFile)

	// Check if file exists and is readable
	stat, err := os.Stat(tokenFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no Vault token found: VAULT_TOKEN environment variable not set and token file does not exist: %s", tokenFile)
		}
		return "", fmt.Errorf("cannot access token file: %w", err)
	}

	// Check file permissions (should be readable only by owner)
	if perm := stat.Mode().Perm(); perm&0077 != 0 {
		return "", fmt.Errorf("token file has insecure permissions %04o (should be 0600 or 0400)", perm)
	}

	// Read token
	token, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", fmt.Errorf("failed to read token file %s: %w", tokenFile, err)
	}

	tokenStr := strings.TrimSpace(string(token))
	if len(tokenStr) == 0 {
		return "", fmt.Errorf("token file is empty: %s", tokenFile)
	}

	return tokenStr, nil
}

// generateBasePath creates a unique base path for the current user/host
func generateBasePath() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}

	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME") // Windows compatibility
	}
	if username == "" {
		username = "unknown-user"
	}

	// Sanitize hostname and username for use in paths
	hostname = sanitizePathComponent(hostname)
	username = sanitizePathComponent(username)

	return fmt.Sprintf("users/%s-%s", hostname, username), nil
}

// sanitizePathComponent removes or replaces characters that could cause issues in Vault paths
func sanitizePathComponent(component string) string {
	// Replace problematic characters with underscores
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
		"\t", "_",
		"\n", "_",
		"\r", "_",
	)

	cleaned := replacer.Replace(component)

	// Ensure it's not empty and doesn't start with special characters
	if len(cleaned) == 0 {
		return "unknown"
	}

	// Remove leading dots or dashes
	cleaned = strings.TrimLeft(cleaned, ".-")
	if len(cleaned) == 0 {
		return "unknown"
	}

	return cleaned
}
