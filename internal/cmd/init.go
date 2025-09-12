package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-vault-keeper/internal/config"
	"github.com/rzago/ssh-vault-keeper/internal/vault"
	"github.com/spf13/cobra"
)

// newInitCommand creates the init command
func newInitCommand(cfg *config.Config) *cobra.Command {
	var (
		vaultAddr     string
		token         string
		mountPath     string
		configPath    string
		forceOverwrite bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize SSH Vault Keeper configuration and Vault setup",
		Long: `Initialize SSH Vault Keeper by:
1. Creating configuration file
2. Setting up Vault token
3. Testing Vault connection
4. Creating necessary Vault mounts and paths`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cfg, initOptions{
				vaultAddr:      vaultAddr,
				token:          token,
				mountPath:      mountPath,
				configPath:     configPath,
				forceOverwrite: forceOverwrite,
			})
		},
	}

	// Command-specific flags
	cmd.Flags().StringVar(&vaultAddr, "vault-addr", cfg.Vault.Address, "Vault server address")
	cmd.Flags().StringVar(&token, "token", "", "Vault authentication token")
	cmd.Flags().StringVar(&mountPath, "mount-path", cfg.Vault.MountPath, "Vault mount path for SSH backups")
	cmd.Flags().StringVar(&configPath, "config-path", "", "Custom configuration file path")
	cmd.Flags().BoolVar(&forceOverwrite, "force", false, "Overwrite existing configuration")

	return cmd
}

type initOptions struct {
	vaultAddr      string
	token          string
	mountPath      string
	configPath     string
	forceOverwrite bool
}

func runInit(cfg *config.Config, opts initOptions) error {
	log.Info().Msg("Starting SSH Vault Keeper initialization")

	// Determine config file path
	configFile := opts.configPath
	if configFile == "" {
		configFile = config.GetConfigPath()
	}

	// Check if config already exists
	if _, err := os.Stat(configFile); err == nil && !opts.forceOverwrite {
		fmt.Printf("Configuration file already exists: %s\n", configFile)
		fmt.Printf("Use --force to overwrite, or specify a different path with --config-path\n")
		return nil
	}

	// Update configuration with provided values
	if opts.vaultAddr != "" {
		cfg.Vault.Address = opts.vaultAddr
	}
	if opts.mountPath != "" {
		cfg.Vault.MountPath = opts.mountPath
	}

	// Handle token setup
	if err := setupToken(cfg, opts.token); err != nil {
		return fmt.Errorf("failed to setup token: %w", err)
	}

	// Test Vault connection
	fmt.Printf("Testing Vault connection to %s...\n", cfg.Vault.Address)
	vaultClient, err := vault.New(&cfg.Vault)
	if err != nil {
		return fmt.Errorf("failed to connect to Vault: %w", err)
	}
	defer vaultClient.Close()

	if err := vaultClient.TestConnection(); err != nil {
		return fmt.Errorf("Vault connection test failed: %w", err)
	}
	fmt.Printf("✓ Vault connection successful\n")

	// Ensure mount exists
	fmt.Printf("Setting up Vault mount: %s\n", cfg.Vault.MountPath)
	if err := vaultClient.EnsureMountExists(); err != nil {
		return fmt.Errorf("failed to setup Vault mount: %w", err)
	}
	fmt.Printf("✓ Vault mount ready\n")

	// Save configuration
	if err := cfg.Save(configFile); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Create SSH directory if it doesn't exist (for first-time users)
	if _, err := os.Stat(cfg.Backup.SSHDir); os.IsNotExist(err) {
		fmt.Printf("Creating SSH directory: %s\n", cfg.Backup.SSHDir)
		if err := os.MkdirAll(cfg.Backup.SSHDir, 0700); err != nil {
			log.Warn().Err(err).Msg("Failed to create SSH directory")
		}
	}

	// Success message
	fmt.Printf("\n✓ SSH Vault Keeper initialized successfully!\n")
	fmt.Printf("\nConfiguration saved to: %s\n", configFile)
	fmt.Printf("Vault server: %s\n", cfg.Vault.Address)
	fmt.Printf("Mount path: %s\n", cfg.Vault.MountPath)
	fmt.Printf("SSH directory: %s\n", cfg.Backup.SSHDir)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Run 'ssh-vault-keeper analyze' to see your SSH files\n")
	fmt.Printf("  2. Run 'ssh-vault-keeper backup' to create your first backup\n")

	return nil
}

// setupToken handles token configuration
func setupToken(cfg *config.Config, token string) error {
	if token == "" {
		// Try to read token from environment
		if envToken := os.Getenv("VAULT_TOKEN"); envToken != "" {
			token = envToken
		} else {
			return fmt.Errorf("no Vault token provided. Use --token flag or VAULT_TOKEN environment variable")
		}
	}

	// Ensure token directory exists
	tokenDir := filepath.Dir(cfg.Vault.TokenFile)
	if err := os.MkdirAll(tokenDir, 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	// Expand home directory in token file path
	tokenFile := cfg.Vault.TokenFile
	if tokenFile[0] == '~' {
		homeDir, _ := os.UserHomeDir()
		tokenFile = filepath.Join(homeDir, tokenFile[2:])
	}

	// Write token file
	if err := os.WriteFile(tokenFile, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	fmt.Printf("✓ Token saved to %s\n", tokenFile)
	return nil
}
