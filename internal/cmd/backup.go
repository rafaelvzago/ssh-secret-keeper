package cmd

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-vault-keeper/internal/config"
	"github.com/rzago/ssh-vault-keeper/internal/ssh"
	"github.com/rzago/ssh-vault-keeper/internal/vault"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// newBackupCommand creates the backup command
func newBackupCommand(cfg *config.Config) *cobra.Command {
	var (
		backupName  string
		sshDir      string
		passphrase  string
		dryRun      bool
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "backup [name]",
		Short: "Backup SSH directory to Vault",
		Long: `Backup your SSH directory to Vault with client-side encryption.
The backup includes all SSH keys, configuration files, and metadata.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Use provided name or generate timestamp-based name
			name := backupName
			if len(args) > 0 {
				name = args[0]
			}
			if name == "" {
				name = fmt.Sprintf("backup-%s", time.Now().Format("20060102-150405"))
			}

			return runBackup(cfg, backupOptions{
				name:        name,
				sshDir:      sshDir,
				passphrase:  passphrase,
				dryRun:      dryRun,
				interactive: interactive,
			})
		},
	}

	// Command-specific flags
	cmd.Flags().StringVar(&backupName, "name", "", "Custom backup name (default: backup-YYYYMMDD-HHMMSS)")
	cmd.Flags().StringVar(&sshDir, "ssh-dir", cfg.Backup.SSHDir, "SSH directory to backup")
	cmd.Flags().StringVar(&passphrase, "passphrase", "", "Encryption passphrase (will prompt if not provided)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be backed up without actually doing it")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactively select files to backup")

	return cmd
}

type backupOptions struct {
	name        string
	sshDir      string
	passphrase  string
	dryRun      bool
	interactive bool
}

func runBackup(cfg *config.Config, opts backupOptions) error {
	log.Info().
		Str("backup_name", opts.name).
		Str("ssh_dir", opts.sshDir).
		Bool("dry_run", opts.dryRun).
		Msg("Starting backup process")

	// Initialize SSH handler
	sshHandler := ssh.New()

	// Read and analyze SSH directory
	fmt.Printf("Analyzing SSH directory: %s\n", opts.sshDir)
	backupData, err := sshHandler.ReadDirectory(opts.sshDir)
	if err != nil {
		return fmt.Errorf("failed to read SSH directory: %w", err)
	}

	// Display analysis summary
	displayBackupSummary(backupData)

	if opts.dryRun {
		fmt.Printf("\n[DRY RUN] Backup '%s' would include %d files\n", opts.name, len(backupData.Files))
		return nil
	}

	// Interactive file selection
	if opts.interactive {
		if err := interactiveFileSelection(backupData); err != nil {
			return fmt.Errorf("interactive selection failed: %w", err)
		}
	}

	// Get encryption passphrase
	passphrase := opts.passphrase
	if passphrase == "" {
		passphrase, err = promptPassphrase("Enter encryption passphrase: ")
		if err != nil {
			return fmt.Errorf("failed to get passphrase: %w", err)
		}
	}

	// Encrypt backup
	fmt.Printf("Encrypting backup data...\n")
	if err := sshHandler.EncryptBackup(backupData, passphrase); err != nil {
		return fmt.Errorf("failed to encrypt backup: %w", err)
	}
	fmt.Printf("✓ Backup encrypted successfully\n")

	// Connect to Vault
	fmt.Printf("Connecting to Vault...\n")
	vaultClient, err := vault.New(&cfg.Vault)
	if err != nil {
		return fmt.Errorf("failed to connect to Vault: %w", err)
	}
	defer vaultClient.Close()

	// Prepare vault data (remove content fields for storage)
	vaultData := prepareVaultData(backupData)

	// Store backup in Vault
	fmt.Printf("Storing backup in Vault...\n")
	if err := vaultClient.StoreBackup(opts.name, vaultData); err != nil {
		return fmt.Errorf("failed to store backup: %w", err)
	}

	// Update metadata
	if err := updateBackupMetadata(vaultClient, opts.name, backupData); err != nil {
		log.Warn().Err(err).Msg("Failed to update metadata")
	}

	fmt.Printf("✓ Backup '%s' completed successfully\n", opts.name)
	fmt.Printf("Files backed up: %d\n", len(backupData.Files))
	fmt.Printf("Total size: %d bytes\n", backupData.Metadata["total_size"])

	// Show permission preservation summary
	fmt.Printf("\nPermission Preservation:\n")
	permissionMap := make(map[string]int)
	for _, fileData := range backupData.Files {
		permStr := fmt.Sprintf("%04o", fileData.Permissions&os.ModePerm)
		permissionMap[permStr]++
	}
	
	for perm, count := range permissionMap {
		fmt.Printf("• %s: %d files\n", perm, count)
	}
	
	fmt.Printf("\n✅ All file permissions have been preserved in the backup\n")

	return nil
}

// displayBackupSummary shows a summary of what will be backed up
func displayBackupSummary(backup *ssh.BackupData) {
	fmt.Printf("\n📋 Backup Analysis Summary\n")
	fmt.Printf("═══════════════════════════\n")
	fmt.Printf("Total files: %d\n", len(backup.Files))
	fmt.Printf("Key pairs: %d\n", backup.Analysis.Summary.KeyPairCount)
	fmt.Printf("Service keys: %d\n", backup.Analysis.Summary.ServiceKeys)
	fmt.Printf("Personal keys: %d\n", backup.Analysis.Summary.PersonalKeys)
	fmt.Printf("Work keys: %d\n", backup.Analysis.Summary.WorkKeys)
	fmt.Printf("System files: %d\n", backup.Analysis.Summary.SystemFiles)

	// Show key pairs
	if len(backup.Analysis.KeyPairs) > 0 {
		fmt.Printf("\n🔑 Key Pairs Found:\n")
		for baseName, keyPair := range backup.Analysis.KeyPairs {
			fmt.Printf("  • %s", baseName)
			if keyPair.PrivateKeyFile != "" && keyPair.PublicKeyFile != "" {
				fmt.Printf(" (complete pair)")
			} else if keyPair.PrivateKeyFile != "" {
				fmt.Printf(" (private key only)")
			} else {
				fmt.Printf(" (public key only)")
			}
			fmt.Printf("\n")
		}
	}

	// Show categories
	if len(backup.Analysis.Categories) > 0 {
		fmt.Printf("\n📁 Categories:\n")
		for category, files := range backup.Analysis.Categories {
			fmt.Printf("  • %s: %d files\n", category, len(files))
		}
	}
}

// interactiveFileSelection allows user to select which files to backup
func interactiveFileSelection(backup *ssh.BackupData) error {
	fmt.Printf("\n🎯 Interactive File Selection\n")
	fmt.Printf("Select files to include in backup (y/n/a=all/q=quit):\n")

	for filename, fileData := range backup.Files {
		fmt.Printf("Include '%s' [%s, %d bytes]? [y/N/a/q]: ", 
			filename, 
			fileData.KeyInfo.Type, 
			fileData.Size)

		var response string
		fmt.Scanln(&response)

		switch response {
		case "y", "Y", "yes":
			// Keep file
		case "a", "A", "all":
			// Keep all remaining files
			return nil
		case "q", "Q", "quit":
			return fmt.Errorf("backup cancelled by user")
		default:
			// Remove file from backup
			delete(backup.Files, filename)
		}
	}

	return nil
}

// promptPassphrase securely prompts for a passphrase
func promptPassphrase(prompt string) (string, error) {
	fmt.Print(prompt)
	passphrase, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	
	if err != nil {
		return "", err
	}
	
	return string(passphrase), nil
}

// prepareVaultData converts backup data for Vault storage
func prepareVaultData(backup *ssh.BackupData) map[string]interface{} {
	data := map[string]interface{}{
		"version":   backup.Version,
		"timestamp": backup.Timestamp.Format(time.RFC3339),
		"hostname":  backup.Hostname,
		"username":  backup.Username,
		"ssh_dir":   backup.SSHDir,
		"metadata":  backup.Metadata,
		"files":     make(map[string]interface{}),
	}

	// Convert file data for storage (without raw content)
	files := make(map[string]interface{})
	for filename, fileData := range backup.Files {
		files[filename] = map[string]interface{}{
			"filename":     fileData.Filename,
			"permissions":  int(fileData.Permissions),
			"size":         fileData.Size,
			"mod_time":     fileData.ModTime.Format(time.RFC3339),
			"checksum":     fileData.Checksum,
			"encrypted":    fileData.Encrypted,
			"key_info":     fileData.KeyInfo,
		}
	}
	data["files"] = files

	return data
}

// updateBackupMetadata updates the backup metadata in Vault
func updateBackupMetadata(client *vault.Client, backupName string, backup *ssh.BackupData) error {
	metadata, err := client.GetMetadata()
	if err != nil {
		return err
	}

	// Initialize backups list if it doesn't exist
	if metadata["backups"] == nil {
		metadata["backups"] = make(map[string]interface{})
	}

	backups := metadata["backups"].(map[string]interface{})
	backups[backupName] = map[string]interface{}{
		"timestamp":   backup.Timestamp.Format(time.RFC3339),
		"file_count":  len(backup.Files),
		"total_size":  backup.Metadata["total_size"],
		"hostname":    backup.Hostname,
		"username":    backup.Username,
	}

	return client.StoreMetadata(metadata)
}
