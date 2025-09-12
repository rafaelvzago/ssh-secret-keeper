package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-vault-keeper/internal/config"
	"github.com/rzago/ssh-vault-keeper/internal/vault"
	"github.com/spf13/cobra"
)

// newStatusCommand creates the status command
func newStatusCommand(cfg *config.Config) *cobra.Command {
	var (
		checkVault bool
		checkSSH   bool
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of SSH Vault Keeper configuration and connections",
		Long: `Check the status of your SSH Vault Keeper setup including:
- Configuration file status
- Vault connection
- SSH directory analysis
- Recent backup information`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cfg, statusOptions{
				checkVault: checkVault,
				checkSSH:   checkSSH,
			})
		},
	}

	// Command-specific flags
	cmd.Flags().BoolVar(&checkVault, "vault", true, "Check Vault connection")
	cmd.Flags().BoolVar(&checkSSH, "ssh", true, "Check SSH directory status")

	return cmd
}

type statusOptions struct {
	checkVault bool
	checkSSH   bool
}

func runStatus(cfg *config.Config, opts statusOptions) error {
	log.Info().
		Bool("check_vault", opts.checkVault).
		Bool("check_ssh", opts.checkSSH).
		Msg("Checking SSH Vault Keeper status")

	fmt.Printf("🔍 SSH Vault Keeper Status\n")
	fmt.Printf("═══════════════════════════\n\n")

	// Configuration status
	fmt.Printf("📋 Configuration:\n")
	fmt.Printf("  Config version: %s\n", cfg.Version)
	fmt.Printf("  SSH directory: %s", cfg.Backup.SSHDir)
	if _, err := os.Stat(cfg.Backup.SSHDir); err != nil {
		fmt.Printf(" ❌ (not found)")
	} else {
		fmt.Printf(" ✅")
	}
	fmt.Printf("\n")

	fmt.Printf("  Vault address: %s\n", cfg.Vault.Address)
	fmt.Printf("  Mount path: %s\n", cfg.Vault.MountPath)
	fmt.Printf("  Token file: %s", cfg.Vault.TokenFile)
	
	// Check token file
	if _, err := os.Stat(cfg.Vault.TokenFile); err != nil {
		fmt.Printf(" ❌ (not found)")
	} else {
		fmt.Printf(" ✅")
	}
	fmt.Printf("\n")

	// Vault connection check
	if opts.checkVault {
		fmt.Printf("\n🔐 Vault Connection:\n")
		vaultClient, err := vault.New(&cfg.Vault)
		if err != nil {
			fmt.Printf("  Connection: ❌ Failed to create client\n")
			fmt.Printf("  Error: %v\n", err)
		} else {
			defer vaultClient.Close()

			if err := vaultClient.TestConnection(); err != nil {
				fmt.Printf("  Connection: ❌ Failed\n")
				fmt.Printf("  Error: %v\n", err)
			} else {
				fmt.Printf("  Connection: ✅ Success\n")
				fmt.Printf("  Base path: %s\n", vaultClient.GetBasePath())

				// Check for existing backups
				backups, err := vaultClient.ListBackups()
				if err != nil {
					fmt.Printf("  Backups: ❌ Failed to list\n")
				} else {
					fmt.Printf("  Backups: %d found\n", len(backups))
					if len(backups) > 0 {
						fmt.Printf("    Most recent: %s\n", backups[len(backups)-1])
					}
				}
			}
		}
	}

	// SSH directory analysis
	if opts.checkSSH {
		fmt.Printf("\n📁 SSH Directory Analysis:\n")
		if _, err := os.Stat(cfg.Backup.SSHDir); err != nil {
			fmt.Printf("  Directory: ❌ %s does not exist\n", cfg.Backup.SSHDir)
		} else {
			// Quick file count
			files, err := os.ReadDir(cfg.Backup.SSHDir)
			if err != nil {
				fmt.Printf("  Directory: ❌ Cannot read directory\n")
			} else {
				fileCount := 0
				for _, file := range files {
					if !file.IsDir() {
						fileCount++
					}
				}
				fmt.Printf("  Directory: ✅ %s\n", cfg.Backup.SSHDir)
				fmt.Printf("  Files found: %d\n", fileCount)

				// Quick analysis of common files
				commonFiles := []string{"id_rsa", "id_rsa.pub", "config", "known_hosts", "authorized_keys"}
				foundFiles := 0
				for _, commonFile := range commonFiles {
					filePath := fmt.Sprintf("%s/%s", cfg.Backup.SSHDir, commonFile)
					if _, err := os.Stat(filePath); err == nil {
						foundFiles++
					}
				}
				fmt.Printf("  Common SSH files: %d/%d found\n", foundFiles, len(commonFiles))

				// Check permissions
				dirInfo, err := os.Stat(cfg.Backup.SSHDir)
				if err == nil {
					perms := dirInfo.Mode().Perm()
					if perms == 0700 {
						fmt.Printf("  Permissions: ✅ %s (secure)\n", perms.String())
					} else {
						fmt.Printf("  Permissions: ⚠️  %s (recommend 700)\n", perms.String())
					}
				}
			}
		}
	}

	// Security settings
	fmt.Printf("\n🔒 Security Settings:\n")
	fmt.Printf("  Encryption algorithm: %s\n", cfg.Security.Algorithm)
	fmt.Printf("  Key derivation: %s (%d iterations)\n", cfg.Security.KeyDerivation, cfg.Security.Iterations)
	fmt.Printf("  Per-file encryption: %v\n", cfg.Security.PerFileEncrypt)
	fmt.Printf("  Integrity verification: %v\n", cfg.Security.VerifyIntegrity)

	// Recommendations
	fmt.Printf("\n💡 Recommendations:\n")
	
	// Check if SSH directory exists and has files
	if _, err := os.Stat(cfg.Backup.SSHDir); err != nil {
		fmt.Printf("  • Create SSH directory: mkdir -p %s && chmod 700 %s\n", cfg.Backup.SSHDir, cfg.Backup.SSHDir)
	} else if files, err := os.ReadDir(cfg.Backup.SSHDir); err == nil {
		fileCount := 0
		for _, file := range files {
			if !file.IsDir() {
				fileCount++
			}
		}
		if fileCount == 0 {
			fmt.Printf("  • SSH directory is empty - consider generating SSH keys\n")
		} else if opts.checkVault {
			if vaultClient, err := vault.New(&cfg.Vault); err == nil {
				if backups, err := vaultClient.ListBackups(); err == nil && len(backups) == 0 {
					fmt.Printf("  • No backups found - run 'ssh-vault-keeper backup' to create one\n")
				}
				vaultClient.Close()
			}
		}
	}

	// Check token file permissions
	if tokenInfo, err := os.Stat(cfg.Vault.TokenFile); err == nil {
		if tokenInfo.Mode().Perm() != 0600 {
			fmt.Printf("  • Fix token file permissions: chmod 600 %s\n", cfg.Vault.TokenFile)
		}
	}

	fmt.Printf("  • Run 'ssh-vault-keeper analyze' to see detailed SSH file analysis\n")
	fmt.Printf("  • Run 'ssh-vault-keeper list' to see available backups\n")

	return nil
}
