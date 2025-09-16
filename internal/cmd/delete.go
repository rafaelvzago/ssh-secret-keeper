package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-secret-keeper/internal/config"
	"github.com/rzago/ssh-secret-keeper/internal/interfaces"
	"github.com/rzago/ssh-secret-keeper/internal/storage"
	"github.com/spf13/cobra"
)

// newDeleteCommand creates the delete command
func newDeleteCommand(cfg *config.Config) *cobra.Command {
	var (
		force       bool
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "delete <backup-name>",
		Short: "Delete a backup from Vault",
		Long: `Delete a backup from Vault storage. This operation is irreversible.
You can delete a specific backup by name, or use interactive mode to select from available backups.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backupName := args[0]
			return runDelete(cfg, deleteOptions{
				backupName:  backupName,
				force:       force,
				interactive: interactive,
			})
		},
	}

	// Command-specific flags
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive backup selection")

	return cmd
}

type deleteOptions struct {
	backupName  string
	force       bool
	interactive bool
}

func runDelete(cfg *config.Config, opts deleteOptions) error {
	log.Info().
		Str("backup_name", opts.backupName).
		Bool("force", opts.force).
		Bool("interactive", opts.interactive).
		Msg("Starting backup deletion")

	// Create storage provider via factory
	factory := storage.NewFactory()
	storageProvider, err := factory.CreateStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to create storage provider: %w", err)
	}
	defer storageProvider.Close()

	ctx := context.Background()

	// Test connection
	fmt.Printf("Connecting to %s storage...\n", storageProvider.GetProviderType())
	if err := storageProvider.TestConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to storage: %w", err)
	}
	fmt.Printf("✓ Connected to %s\n", storageProvider.GetProviderType())

	// Handle interactive mode
	if opts.interactive {
		backupName, err := interactiveBackupSelection(storageProvider)
		if err != nil {
			return fmt.Errorf("interactive selection failed: %w", err)
		}
		if backupName == "" {
			fmt.Printf("No backup selected, operation cancelled\n")
			return nil
		}
		opts.backupName = backupName
	}

	// Verify backup exists
	fmt.Printf("Checking if backup '%s' exists...\n", opts.backupName)
	backups, err := storageProvider.ListBackups(ctx)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	backupExists := false
	for _, backup := range backups {
		if backup == opts.backupName {
			backupExists = true
			break
		}
	}

	if !backupExists {
		return fmt.Errorf("backup '%s' not found. Available backups: %s",
			opts.backupName, strings.Join(backups, ", "))
	}

	// Show backup details before deletion
	fmt.Printf("✓ Backup '%s' found\n", opts.backupName)
	if err := showBackupDetailsForDeletion(storageProvider, opts.backupName); err != nil {
		log.Warn().Err(err).Msg("Failed to get backup details")
		// Continue with deletion even if we can't show details
	}

	// Confirmation prompt (unless --force is used)
	if !opts.force {
		fmt.Printf("\n⚠️  WARNING: This will permanently delete backup '%s'\n", opts.backupName)
		fmt.Printf("This operation cannot be undone!\n")
		fmt.Printf("\nAre you sure you want to delete this backup? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			fmt.Printf("Deletion cancelled\n")
			return nil
		}
	}

	// Delete the backup
	fmt.Printf("Deleting backup '%s'...\n", opts.backupName)
	if err := storageProvider.DeleteBackup(ctx, opts.backupName); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	fmt.Printf("✓ Backup '%s' deleted successfully\n", opts.backupName)

	// Update metadata to remove the backup entry
	if err := updateMetadataAfterDeletion(storageProvider, opts.backupName); err != nil {
		log.Warn().Err(err).Msg("Failed to update metadata after deletion")
	}

	// The ForceDeleteBackup method already handles verification with retries
	// Just provide a final status message
	fmt.Printf("✓ Deletion process completed\n")

	return nil
}

// interactiveBackupSelection allows user to select a backup to delete
func interactiveBackupSelection(provider interfaces.StorageProvider) (string, error) {
	ctx := context.Background()
	backups, err := provider.ListBackups(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list backups: %w", err)
	}

	if len(backups) == 0 {
		fmt.Printf("No backups found to delete\n")
		return "", nil
	}

	fmt.Printf("\n🗑️  Interactive Backup Deletion\n")
	fmt.Printf("Select a backup to delete:\n\n")

	for i, backup := range backups {
		fmt.Printf("  %d. %s\n", i+1, backup)
	}

	fmt.Printf("\nEnter backup number (1-%d) or 'q' to quit: ", len(backups))

	var input string
	fmt.Scanln(&input)

	input = strings.ToLower(strings.TrimSpace(input))
	if input == "q" || input == "quit" {
		return "", nil
	}

	// Parse selection
	var selection int
	if _, err := fmt.Sscanf(input, "%d", &selection); err != nil {
		return "", fmt.Errorf("invalid selection: %s", input)
	}

	if selection < 1 || selection > len(backups) {
		return "", fmt.Errorf("selection out of range: %d", selection)
	}

	return backups[selection-1], nil
}

// showBackupDetailsForDeletion displays backup details before deletion
func showBackupDetailsForDeletion(provider interfaces.StorageProvider, backupName string) error {
	ctx := context.Background()
	backupData, err := provider.GetBackup(ctx, backupName)
	if err != nil {
		return err
	}

	// Parse backup data - handle different possible structures
	var files map[string]interface{}
	if filesData, ok := backupData["files"].(map[string]interface{}); ok {
		files = filesData
	} else {
		// If files is not directly available, try to get basic info
		fmt.Printf("\n📁 Backup Details:\n")
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

		// Try to get file count from metadata
		if metadata, ok := backupData["metadata"].(map[string]interface{}); ok {
			if fileCount, ok := metadata["total_files"].(float64); ok {
				fmt.Printf("Files: %d\n", int(fileCount))
			}
		}

		return nil
	}

	fmt.Printf("\n📁 Backup Details:\n")
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

	// Show file summary
	fileTypes := make(map[string]int)
	for _, fileInfo := range files {
		if info, ok := fileInfo.(map[string]interface{}); ok {
			if keyInfo, ok := info["key_info"].(map[string]interface{}); ok {
				if fileType, ok := keyInfo["type"].(string); ok {
					fileTypes[fileType]++
				}
			}
		}
	}

	if len(fileTypes) > 0 {
		fmt.Printf("\nFile Types:\n")
		for fileType, count := range fileTypes {
			fmt.Printf("  • %s: %d files\n", fileType, count)
		}
	}

	return nil
}

// updateMetadataAfterDeletion removes the backup from metadata
func updateMetadataAfterDeletion(provider interfaces.StorageProvider, backupName string) error {
	ctx := context.Background()
	metadata, err := provider.GetMetadata(ctx)
	if err != nil {
		return err
	}

	// Remove backup from metadata
	if backups, ok := metadata["backups"].(map[string]interface{}); ok {
		delete(backups, backupName)
		metadata["backups"] = backups

		// Store updated metadata
		if err := provider.StoreMetadata(ctx, metadata); err != nil {
			return err
		}

		log.Debug().Str("backup", backupName).Msg("Removed backup from metadata")
	}

	return nil
}
