package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-vault-keeper/internal/config"
	"github.com/rzago/ssh-vault-keeper/internal/vault"
	"github.com/spf13/cobra"
)

// newStatusCommand creates the status command
func newStatusCommand(cfg *config.Config) *cobra.Command {
	var (
		checkVault    bool
		checkSSH      bool
		showChecksums bool
		backupName    string
	)

	cmd := &cobra.Command{
		Use:   "status [backup-name]",
		Short: "Show status of SSH Vault Keeper configuration and connections",
		Long: `Check the status of your SSH Vault Keeper setup including:
- Configuration file status
- Vault connection
- SSH directory analysis
- Recent backup information
- File MD5 checksums (with --checksums flag)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				backupName = args[0]
			}
			return runStatus(cfg, statusOptions{
				checkVault:    checkVault,
				checkSSH:      checkSSH,
				showChecksums: showChecksums,
				backupName:    backupName,
			})
		},
	}

	// Command-specific flags
	cmd.Flags().BoolVar(&checkVault, "vault", true, "Check Vault connection")
	cmd.Flags().BoolVar(&checkSSH, "ssh", true, "Check SSH directory status")
	cmd.Flags().BoolVar(&showChecksums, "checksums", false, "Show MD5 checksums for backup files")
	cmd.Flags().StringVar(&backupName, "backup", "", "Show detailed info for specific backup")

	return cmd
}

type statusOptions struct {
	checkVault    bool
	checkSSH      bool
	showChecksums bool
	backupName    string
}

func runStatus(cfg *config.Config, opts statusOptions) error {
	log.Info().
		Bool("check_vault", opts.checkVault).
		Bool("check_ssh", opts.checkSSH).
		Bool("show_checksums", opts.showChecksums).
		Str("backup_name", opts.backupName).
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

					// Show detailed backup info if specific backup requested
					if opts.backupName != "" {
						if err := showBackupDetails(vaultClient, opts.backupName, opts.showChecksums); err != nil {
							fmt.Printf("  ❌ Failed to get backup details: %v\n", err)
						}
					} else if opts.showChecksums && len(backups) > 0 {
						// Show checksums for most recent backup
						mostRecent := backups[len(backups)-1]
						fmt.Printf("\n📋 Most Recent Backup Details (%s):\n", mostRecent)
						if err := showBackupDetails(vaultClient, mostRecent, true); err != nil {
							fmt.Printf("  ❌ Failed to get backup details: %v\n", err)
						}
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
	if opts.showChecksums {
		fmt.Printf("  • Use 'ssh-vault-keeper status --checksums' to view MD5 checksums\n")
	}

	return nil
}

// showBackupDetails displays detailed information about a specific backup
func showBackupDetails(vaultClient *vault.Client, backupName string, showChecksums bool) error {
	backupData, err := vaultClient.GetBackup(backupName)
	if err != nil {
		return err
	}

	// Parse backup data
	files, ok := backupData["files"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid backup format")
	}

	fmt.Printf("\n📁 Backup Details: %s\n", backupName)
	fmt.Printf("────────────────────────────────────\n")

	if timestamp, ok := backupData["timestamp"].(string); ok {
		fmt.Printf("Timestamp: %s\n", timestamp)
	}

	if hostname, ok := backupData["hostname"].(string); ok {
		fmt.Printf("Hostname: %s\n", hostname)
	}

	if username, ok := backupData["username"].(string); ok {
		fmt.Printf("Username: %s\n", username)
	}

	fmt.Printf("Files: %d\n", len(files))

	if showChecksums {
		fmt.Printf("\n🔐 File MD5 Checksums:\n")
		fmt.Printf("────────────────────────────────────\n")

		// Sort filenames for consistent display
		fileNames := make([]string, 0, len(files))
		for filename := range files {
			fileNames = append(fileNames, filename)
		}
		sort.Strings(fileNames)

		for _, filename := range fileNames {
			fileInfo, ok := files[filename].(map[string]interface{})
			if !ok {
				continue
			}

			var checksum, keyType, permissions string
			var size int64

			if md5Hash, ok := fileInfo["checksum"].(string); ok {
				checksum = md5Hash
			}

			// Try to get type from key_info first, then from top level
			if keyInfo, ok := fileInfo["key_info"].(map[string]interface{}); ok {
				if t, ok := keyInfo["type"].(string); ok {
					keyType = t
				}
			}
			// If no type found in key_info, try top level
			if keyType == "" {
				if t, ok := fileInfo["type"].(string); ok {
					keyType = t
				}
			}

			if perms, ok := fileInfo["permissions"].(json.Number); ok {
				if permInt, err := perms.Int64(); err == nil {
					permissions = fmt.Sprintf("%04o", int(permInt)&0777)
				} else {
					permissions = "unknown"
				}
			} else if perms, ok := fileInfo["permissions"].(float64); ok {
				permissions = fmt.Sprintf("%04o", int(perms)&0777)
			} else if perms, ok := fileInfo["permissions"].(int); ok {
				permissions = fmt.Sprintf("%04o", perms&0777)
			} else if perms, ok := fileInfo["permissions"].(int64); ok {
				permissions = fmt.Sprintf("%04o", int(perms)&0777)
			} else {
				permissions = "unknown"
			}

			if s, ok := fileInfo["size"].(json.Number); ok {
				if sizeInt, err := s.Int64(); err == nil {
					size = sizeInt
				} else {
					size = 0
				}
			} else if s, ok := fileInfo["size"].(float64); ok {
				size = int64(s)
			} else if s, ok := fileInfo["size"].(int); ok {
				size = int64(s)
			} else if s, ok := fileInfo["size"].(int64); ok {
				size = s
			}

			fmt.Printf("  📄 %s\n", filename)
			fmt.Printf("     MD5: %s\n", checksum)
			fmt.Printf("     Type: %s | Size: %d bytes | Permissions: %s\n",
				keyType, size, permissions)
			fmt.Printf("\n")
		}
	}

	return nil
}
