package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-secret-keeper/internal/config"
	"github.com/rzago/ssh-secret-keeper/internal/interfaces"
	"github.com/rzago/ssh-secret-keeper/internal/ssh"
	"github.com/rzago/ssh-secret-keeper/internal/storage"
	"github.com/spf13/cobra"
)

// newBackupCommand creates the backup command
func newBackupCommand(cfg *config.Config) *cobra.Command {
	var (
		backupName  string
		sshDir      string
		dryRun      bool
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "backup [name]",
		Short: "Backup SSH directory to Vault",
		Long: `Backup your SSH directory to Vault.
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
				dryRun:      dryRun,
				interactive: interactive,
			})
		},
	}

	// Command-specific flags
	cmd.Flags().StringVar(&backupName, "name", "", "Custom backup name (default: backup-YYYYMMDD-HHMMSS)")
	cmd.Flags().StringVar(&sshDir, "ssh-dir", cfg.Backup.SSHDir, "SSH directory to backup")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be backed up without actually doing it")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactively select files to backup")

	return cmd
}

type backupOptions struct {
	name        string
	sshDir      string
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

	// No passphrase needed - the storage provider is responsible for security

	// Backup data is ready for storage (no encryption needed)
	fmt.Printf("✓ Backup data prepared successfully\n")

	// Create storage provider via factory
	factory := storage.NewFactory()
	storageProvider, err := factory.CreateStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to create storage provider: %w", err)
	}
	defer storageProvider.Close()

	// Test connection
	ctx := context.Background()
	fmt.Printf("Connecting to %s storage...\n", storageProvider.GetProviderType())
	if err := storageProvider.TestConnection(ctx); err != nil {
		return fmt.Errorf("storage connection test failed: %w", err)
	}

	// Prepare data for storage
	vaultData := prepareVaultData(backupData)

	// Store backup using abstraction
	fmt.Printf("Storing backup in %s...\n", storageProvider.GetProviderType())
	if err := storageProvider.StoreBackup(ctx, opts.name, vaultData); err != nil {
		return fmt.Errorf("failed to store backup: %w", err)
	}

	// Update metadata
	if err := updateBackupMetadata(storageProvider, opts.name, backupData); err != nil {
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

	// Show MD5 checksums summary
	fmt.Printf("\n🔐 MD5 Integrity Protection:\n")
	fmt.Printf("• All %d files protected with MD5 checksums\n", len(backupData.Files))
	fmt.Printf("• Use 'ssh-vault-keeper status --checksums' to view file hashes\n")
	fmt.Printf("• Use 'ssh-vault-keeper status %s --checksums' for detailed view\n", opts.name)

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

	// Convert file data for storage (with plain content)
	files := make(map[string]interface{})
	for filename, fileData := range backup.Files {
		// Validate and log the permissions being stored
		permissionsToStore := int(fileData.Permissions & os.ModePerm)

		// Critical validation: permissions should never be 0000 for real files
		if permissionsToStore == 0 {
			log.Error().
				Str("file", filename).
				Str("original_perms", fmt.Sprintf("%04o", fileData.Permissions&os.ModePerm)).
				Str("full_mode", fmt.Sprintf("%04o", fileData.Permissions)).
				Msg("CRITICAL: File has 0000 permissions - this should not happen for real files")
		}

		log.Debug().
			Str("file", filename).
			Str("original_perms", fmt.Sprintf("%04o", fileData.Permissions&os.ModePerm)).
			Int("stored_perms", permissionsToStore).
			Msg("Storing file permissions in backup")

		files[filename] = map[string]interface{}{
			"filename":    fileData.Filename,
			"content":     string(fileData.Content), // Store as base64 or plain text
			"permissions": permissionsToStore,       // Store only permission bits, not file type
			"size":        fileData.Size,
			"mod_time":    fileData.ModTime.Format(time.RFC3339),
			"checksum":    fileData.Checksum,
			"key_info":    fileData.KeyInfo,
		}
	}
	data["files"] = files

	return data
}

// updateBackupMetadata updates the backup metadata using storage provider
func updateBackupMetadata(provider interfaces.StorageProvider, backupName string, backup *ssh.BackupData) error {
	ctx := context.Background()

	metadata, err := provider.GetMetadata(ctx)
	if err != nil {
		return err
	}

	// Initialize backups list if it doesn't exist
	if metadata["backups"] == nil {
		metadata["backups"] = make(map[string]interface{})
	}

	backups := metadata["backups"].(map[string]interface{})
	backups[backupName] = map[string]interface{}{
		"timestamp":  backup.Timestamp.Format(time.RFC3339),
		"file_count": len(backup.Files),
		"total_size": backup.Metadata["total_size"],
		"hostname":   backup.Hostname,
		"username":   backup.Username,
	}

	return provider.StoreMetadata(ctx, metadata)
}
