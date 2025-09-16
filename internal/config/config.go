package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Version   string         `yaml:"version" mapstructure:"version"`
	Storage   StorageConfig  `yaml:"storage" mapstructure:"storage"` // New multi-provider storage config
	Vault     VaultConfig    `yaml:"vault" mapstructure:"vault"`     // Keep for backward compatibility
	Backup    BackupConfig   `yaml:"backup" mapstructure:"backup"`
	Security  SecurityConfig `yaml:"security" mapstructure:"security"`
	Logging   LoggingConfig  `yaml:"logging" mapstructure:"logging"`
	Detectors DetectorConfig `yaml:"detectors" mapstructure:"detectors"`
}

// VaultConfig holds Vault connection settings
type VaultConfig struct {
	Address       string `yaml:"address" mapstructure:"address"`
	TokenFile     string `yaml:"token_file" mapstructure:"token_file"`
	MountPath     string `yaml:"mount_path" mapstructure:"mount_path"`
	Namespace     string `yaml:"namespace,omitempty" mapstructure:"namespace"`
	TLSSkipVerify bool   `yaml:"tls_skip_verify" mapstructure:"tls_skip_verify"`
}

// BackupConfig holds backup behavior settings
type BackupConfig struct {
	SSHDir          string              `yaml:"ssh_dir" mapstructure:"ssh_dir"`
	HostnamePrefix  bool                `yaml:"hostname_prefix" mapstructure:"hostname_prefix"`
	RetentionCount  int                 `yaml:"retention_count" mapstructure:"retention_count"`
	IncludePatterns []string            `yaml:"include_patterns" mapstructure:"include_patterns"`
	ExcludePatterns []string            `yaml:"exclude_patterns" mapstructure:"exclude_patterns"`
	Categories      map[string][]string `yaml:"categories" mapstructure:"categories"`
}

// SecurityConfig holds encryption and security settings
type SecurityConfig struct {
	Algorithm       string `yaml:"algorithm" mapstructure:"algorithm"`
	KeyDerivation   string `yaml:"key_derivation" mapstructure:"key_derivation"`
	Iterations      int    `yaml:"iterations" mapstructure:"iterations"`
	PerFileEncrypt  bool   `yaml:"per_file_encrypt" mapstructure:"per_file_encrypt"`
	VerifyIntegrity bool   `yaml:"verify_integrity" mapstructure:"verify_integrity"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level  string `yaml:"level" mapstructure:"level"`
	Format string `yaml:"format" mapstructure:"format"`
}

// DetectorConfig holds key detection settings
type DetectorConfig struct {
	Enabled        []string          `yaml:"enabled" mapstructure:"enabled"`
	CustomPatterns string            `yaml:"custom_patterns_file" mapstructure:"custom_patterns_file"`
	ServiceMapping map[string]string `yaml:"service_mapping" mapstructure:"service_mapping"`
	PurposeRules   map[string]string `yaml:"purpose_rules" mapstructure:"purpose_rules"`
}

// Default returns a configuration with default values
func Default() *Config {
	homeDir, _ := os.UserHomeDir()

	return &Config{
		Version: "1.0",
		Storage: StorageConfig{
			Provider: "vault", // Default to vault for backward compatibility
			Vault: &VaultConfig{
				Address:       "http://localhost:8200",
				TokenFile:     filepath.Join(homeDir, ".ssh-vault-keeper", "token"),
				MountPath:     "ssh-backups",
				TLSSkipVerify: false,
			},
		},
		Vault: VaultConfig{
			Address:       "http://localhost:8200", // Default Vault server address - override with VAULT_ADDR env var
			TokenFile:     filepath.Join(homeDir, ".ssh-vault-keeper", "token"),
			MountPath:     "ssh-backups",
			TLSSkipVerify: false,
		},
		Backup: BackupConfig{
			SSHDir:         filepath.Join(homeDir, ".ssh"),
			HostnamePrefix: true,
			RetentionCount: 10,
			IncludePatterns: []string{
				"*.rsa", "*.pem", "*.pub", "id_rsa*",
				"config", "known_hosts*", "authorized_keys",
			},
			ExcludePatterns: []string{"*.tmp", "*.bak", "*.old"},
			Categories: map[string][]string{
				"services": {"service1_rsa", "service2_rsa", "service3_rsa", "service4_rsa"},
				"personal": {"id_rsa", "local_rsa"},
				"work":     {"work_key1.rsa", "work_key2.rsa"},
			},
		},
		Security: SecurityConfig{
			Algorithm:       "AES-256-GCM",
			KeyDerivation:   "PBKDF2",
			Iterations:      100000,
			PerFileEncrypt:  true,
			VerifyIntegrity: true,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "console",
		},
		Detectors: DetectorConfig{
			Enabled: []string{"rsa", "pem", "openssh", "config", "hosts"},
			ServiceMapping: map[string]string{
				"github":    "git_hosting",
				"gitlab":    "git_hosting",
				"bitbucket": "git_hosting",
				"argocd":    "automation",
			},
			PurposeRules: map[string]string{
				"*work*":     "work",
				"*corp*":     "work",
				"*personal*": "personal",
			},
		},
	}
}

// Load loads configuration from file and environment
func Load() (*Config, error) {
	cfg := Default()

	// Setup viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.ssh-vault-keeper")
	viper.AddConfigPath("/etc/ssh-vault-keeper")

	// Environment variables
	viper.SetEnvPrefix("SSH_VAULT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK, use defaults
	}

	// Unmarshal into struct
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Require VAULT_ADDR environment variable to be set
	// This follows HashiCorp Vault's standard environment variable convention
	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		return nil, fmt.Errorf("VAULT_ADDR environment variable is required but not set")
	}
	cfg.Vault.Address = vaultAddr

	// Override token file path if SSH_SECRET_VAULT_TOKEN_FILE is set
	if tokenFileEnv := os.Getenv("SSH_SECRET_VAULT_TOKEN_FILE"); tokenFileEnv != "" {
		cfg.Vault.TokenFile = tokenFileEnv
	}

	return cfg, nil
}

// Save saves the configuration to file
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the default config file path
func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".ssh-vault-keeper", "config.yaml")
}
