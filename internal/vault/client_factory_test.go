package vault

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rzago/ssh-secret-keeper/internal/config"
)

func TestCreateVaultClient(t *testing.T) {
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
			},
			envToken:  "test-token-123",
			wantError: false,
		},
		{
			name: "config with TLS skip verify",
			cfg: &config.VaultConfig{
				Address:       "https://vault.example.com",
				MountPath:     "ssh-backups",
				TLSSkipVerify: true,
			},
			envToken:  "test-token-123",
			wantError: false,
		},
		{
			name: "config with namespace",
			cfg: &config.VaultConfig{
				Address:   "http://localhost:8200",
				MountPath: "ssh-backups",
				Namespace: "test-namespace",
			},
			envToken:  "test-token-123",
			wantError: false,
		},
		{
			name: "empty token should fail",
			cfg: &config.VaultConfig{
				Address:   "http://localhost:8200",
				MountPath: "ssh-backups",
			},
			envToken:  "",
			wantError: true,
		},
		{
			name: "whitespace-only token should fail",
			cfg: &config.VaultConfig{
				Address:   "http://localhost:8200",
				MountPath: "ssh-backups",
			},
			envToken:  "   \t\n   ",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			originalToken := os.Getenv("VAULT_TOKEN")
			defer os.Setenv("VAULT_TOKEN", originalToken)

			if tt.envToken != "" {
				os.Setenv("VAULT_TOKEN", tt.envToken)
			} else {
				os.Unsetenv("VAULT_TOKEN")
			}

			// Set a dummy token file that doesn't exist to ensure we use env token
			tt.cfg.TokenFile = "/nonexistent/token"

			client, err := createVaultClient(tt.cfg)

			if tt.wantError {
				if err == nil {
					t.Errorf("createVaultClient() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("createVaultClient() unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("createVaultClient() returned nil client")
				return
			}

			// Verify client configuration
			if client.Address() != tt.cfg.Address {
				t.Errorf("client address = %s, want %s", client.Address(), tt.cfg.Address)
			}

			if client.Token() != tt.envToken {
				t.Errorf("client token = %s, want %s", client.Token(), tt.envToken)
			}

			if tt.cfg.Namespace != "" {
				if client.Headers().Get("X-Vault-Namespace") != tt.cfg.Namespace {
					t.Errorf("client namespace = %s, want %s",
						client.Headers().Get("X-Vault-Namespace"), tt.cfg.Namespace)
				}
			}
		})
	}
}

func TestLoadToken(t *testing.T) {
	tests := []struct {
		name        string
		envToken    string
		fileContent string
		filePerms   os.FileMode
		wantError   bool
		wantToken   string
	}{
		{
			name:      "env token takes precedence",
			envToken:  "env-token-123",
			wantError: false,
			wantToken: "env-token-123",
		},
		{
			name:        "file token when no env token",
			envToken:    "",
			fileContent: "file-token-456",
			filePerms:   0600,
			wantError:   false,
			wantToken:   "file-token-456",
		},
		{
			name:        "file token with whitespace",
			envToken:    "",
			fileContent: "\n  file-token-789  \n",
			filePerms:   0600,
			wantError:   false,
			wantToken:   "file-token-789",
		},
		{
			name:        "insecure file permissions",
			envToken:    "",
			fileContent: "token-content",
			filePerms:   0644,
			wantError:   true,
		},
		{
			name:        "empty file content",
			envToken:    "",
			fileContent: "   \n  \t  \n",
			filePerms:   0600,
			wantError:   true,
		},
		{
			name:      "no env token and no file",
			envToken:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			originalToken := os.Getenv("VAULT_TOKEN")
			defer os.Setenv("VAULT_TOKEN", originalToken)

			if tt.envToken != "" {
				os.Setenv("VAULT_TOKEN", tt.envToken)
			} else {
				os.Unsetenv("VAULT_TOKEN")
			}

			var tokenFile string
			if tt.fileContent != "" {
				// Create temporary file
				tmpDir := t.TempDir()
				tokenFile = filepath.Join(tmpDir, "token")

				err := os.WriteFile(tokenFile, []byte(tt.fileContent), tt.filePerms)
				if err != nil {
					t.Fatalf("Failed to create test token file: %v", err)
				}
			} else {
				tokenFile = "/nonexistent/token/file"
			}

			token, err := loadToken(tokenFile)

			if tt.wantError {
				if err == nil {
					t.Errorf("loadToken() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("loadToken() unexpected error: %v", err)
				return
			}

			if token != tt.wantToken {
				t.Errorf("loadToken() = %s, want %s", token, tt.wantToken)
			}
		})
	}
}

func TestLoadTokenHomeDirectory(t *testing.T) {
	// Test home directory expansion
	originalToken := os.Getenv("VAULT_TOKEN")
	defer os.Setenv("VAULT_TOKEN", originalToken)
	os.Unsetenv("VAULT_TOKEN")

	// Create temporary home directory
	tmpHome := t.TempDir()
	vaultDir := filepath.Join(tmpHome, ".vault")
	err := os.MkdirAll(vaultDir, 0700)
	if err != nil {
		t.Fatalf("Failed to create vault dir: %v", err)
	}

	tokenFile := filepath.Join(vaultDir, "token")
	err = os.WriteFile(tokenFile, []byte("home-token-123"), 0600)
	if err != nil {
		t.Fatalf("Failed to create token file: %v", err)
	}

	// Mock home directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpHome)

	token, err := loadToken("~/.vault/token")
	if err != nil {
		t.Errorf("loadToken() unexpected error: %v", err)
		return
	}

	if token != "home-token-123" {
		t.Errorf("loadToken() = %s, want home-token-123", token)
	}
}

func TestGenerateBasePath(t *testing.T) {
	// Save original environment
	originalUser := os.Getenv("USER")
	originalUsername := os.Getenv("USERNAME")
	defer func() {
		os.Setenv("USER", originalUser)
		os.Setenv("USERNAME", originalUsername)
	}()

	tests := []struct {
		name         string
		user         string
		username     string
		wantContains string
	}{
		{
			name:         "with USER env var",
			user:         "testuser",
			username:     "",
			wantContains: "testuser",
		},
		{
			name:         "with USERNAME env var (Windows)",
			user:         "",
			username:     "windowsuser",
			wantContains: "windowsuser",
		},
		{
			name:         "no user env vars",
			user:         "",
			username:     "",
			wantContains: "unknown-user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.user != "" {
				os.Setenv("USER", tt.user)
			} else {
				os.Unsetenv("USER")
			}

			if tt.username != "" {
				os.Setenv("USERNAME", tt.username)
			} else {
				os.Unsetenv("USERNAME")
			}

			basePath, err := generateBasePath()
			if err != nil {
				t.Errorf("generateBasePath() unexpected error: %v", err)
				return
			}

			if basePath == "" {
				t.Error("generateBasePath() returned empty string")
				return
			}

			if !contains(basePath, tt.wantContains) {
				t.Errorf("generateBasePath() = %s, should contain %s", basePath, tt.wantContains)
			}

			if !contains(basePath, "users/") {
				t.Errorf("generateBasePath() = %s, should start with users/", basePath)
			}
		})
	}
}

func TestSanitizePathComponent(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal", "normal"},
		{"with spaces", "with_spaces"},
		{"with/slash", "with_slash"},
		{"with\\backslash", "with_backslash"},
		{"with:colon", "with_colon"},
		{"with*star", "with_star"},
		{"with?question", "with_question"},
		{"with\"quote", "with_quote"},
		{"with<less", "with_less"},
		{"with>greater", "with_greater"},
		{"with|pipe", "with_pipe"},
		{"with\ttab", "with_tab"},
		{"with\nnewline", "with_newline"},
		{"with\rcarriage", "with_carriage"},
		{"", "unknown"},
		{"...", "unknown"},
		{"---", "unknown"},
		{".hidden", "hidden"},
		{"-dash", "dash"},
		{"normal-name", "normal-name"},
		{"normal.name", "normal.name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizePathComponent(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizePathComponent(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		s[0:len(substr)] == substr || (len(s) > len(substr) && contains(s[1:], substr)))
}
