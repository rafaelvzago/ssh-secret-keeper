package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-secret-keeper/internal/analyzer"
	"github.com/rzago/ssh-secret-keeper/internal/config"
	"github.com/rzago/ssh-secret-keeper/internal/interfaces"
	"github.com/rzago/ssh-secret-keeper/internal/ssh"
	"github.com/rzago/ssh-secret-keeper/internal/storage"
	"github.com/spf13/cobra"
)

// newRestoreCommand creates the restore command
func newRestoreCommand(cfg *config.Config) *cobra.Command {
	var (
		backupName   string
		targetDir    string
		dryRun       bool
		overwrite    bool
		interactive  bool
		selectBackup bool
		fileFilter   []string
	)

	cmd := &cobra.Command{
		Use:   "restore [backup-name]",
		Short: "Restore SSH backup from Vault",
		Long: `Restore SSH files from a Vault backup to your SSH directory.
If no backup name is provided, the most recent backup will be used.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := backupName
			if len(args) > 0 {
				name = args[0]
			}

			return runRestore(cfg, restoreOptions{
				backupName:   name,
				targetDir:    targetDir,
				dryRun:       dryRun,
				overwrite:    overwrite,
				interactive:  interactive,
				selectBackup: selectBackup,
				fileFilter:   fileFilter,
			})
		},
	}

	// Command-specific flags
	cmd.Flags().StringVar(&backupName, "backup", "", "Backup name to restore (default: most recent)")
	cmd.Flags().StringVar(&targetDir, "target-dir", cfg.Backup.SSHDir, "Target directory for restored files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be restored without actually doing it")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing files without asking")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactively select files to restore")
	cmd.Flags().BoolVar(&selectBackup, "select", false, "Interactively select which backup to restore")
	cmd.Flags().StringSliceVar(&fileFilter, "files", []string{}, "Only restore specific files (glob patterns)")

	return cmd
}

type restoreOptions struct {
	backupName   string
	targetDir    string
	dryRun       bool
	overwrite    bool
	interactive  bool
	selectBackup bool
	fileFilter   []string
}

func runRestore(cfg *config.Config, opts restoreOptions) error {
	log.Info().
		Str("backup_name", opts.backupName).
		Str("target_dir", opts.targetDir).
		Bool("dry_run", opts.dryRun).
		Msg("Starting restore process")

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

	// Determine backup name
	backupName := opts.backupName
	if backupName == "" {
		if opts.selectBackup {
			backupName, err = selectBackupInteractively(storageProvider)
			if err != nil {
				return fmt.Errorf("failed to select backup: %w", err)
			}
		} else {
			backupName, err = getLatestBackupName(storageProvider)
			if err != nil {
				return fmt.Errorf("failed to find latest backup: %w", err)
			}
			fmt.Printf("Using most recent backup: %s\n", backupName)
		}
	}

	// Retrieve backup from storage
	fmt.Printf("Retrieving backup '%s' from %s...\n", backupName, storageProvider.GetProviderType())
	vaultData, err := storageProvider.GetBackup(ctx, backupName)
	if err != nil {
		return fmt.Errorf("failed to retrieve backup: %w", err)
	}

	// Convert vault data back to backup structure
	backupData, err := parseVaultBackup(vaultData)
	if err != nil {
		return fmt.Errorf("failed to parse backup data: %w", err)
	}

	// Display restore summary
	displayRestoreSummary(backupData, opts.targetDir)

	if opts.dryRun {
		fmt.Printf("\n[DRY RUN] Would restore %d files to %s\n", len(backupData.Files), opts.targetDir)
		return nil
	}

	// No passphrase needed - Vault provides the security

	// Initialize SSH handler
	sshHandler := ssh.New()

	// Backup data is ready for restore (no decryption needed)
	fmt.Printf("✓ Backup data loaded successfully\n")

	// Verify backup integrity
	if cfg.Security.VerifyIntegrity {
		fmt.Printf("Verifying backup integrity...\n")
		if err := sshHandler.VerifyBackup(backupData); err != nil {
			return fmt.Errorf("backup integrity check failed: %w", err)
		}
		fmt.Printf("✓ Backup integrity verified\n")
	}

	// Interactive file selection
	if opts.interactive {
		if err := interactiveRestoreSelection(backupData); err != nil {
			return fmt.Errorf("interactive selection failed: %w", err)
		}
	}

	// Set up restore options
	restoreOpts := ssh.RestoreOptions{
		DryRun:      opts.dryRun,
		Overwrite:   opts.overwrite,
		Interactive: opts.interactive && !opts.overwrite,
		FileFilter:  opts.fileFilter,
	}

	// Restore files
	fmt.Printf("Restoring files to %s...\n", opts.targetDir)
	if err := sshHandler.RestoreFiles(backupData, opts.targetDir, restoreOpts); err != nil {
		return fmt.Errorf("failed to restore files: %w", err)
	}

	// Verify restored permissions
	if !opts.dryRun {
		fmt.Printf("Verifying file permissions...\n")
		if err := sshHandler.VerifyRestorePermissions(backupData, opts.targetDir); err != nil {
			log.Warn().Err(err).Msg("Permission verification completed with warnings")
			fmt.Printf("⚠️  Permission verification completed with warnings (check logs)\n")
		} else {
			fmt.Printf("✓ All file permissions verified\n")
		}
	}

	fmt.Printf("✓ Restore completed successfully\n")
	fmt.Printf("Files restored: %d\n", len(backupData.Files))

	// Show permission summary
	fmt.Printf("\n📋 Permission Summary:\n")
	fmt.Printf("• SSH directory: %s (0700)\n", opts.targetDir)

	privateKeyCount := 0
	publicKeyCount := 0
	for _, fileData := range backupData.Files {
		if fileData.KeyInfo != nil {
			switch fileData.KeyInfo.Type {
			case analyzer.KeyTypePrivate:
				privateKeyCount++
			case analyzer.KeyTypePublic:
				publicKeyCount++
			}
		}
	}

	if privateKeyCount > 0 {
		fmt.Printf("• Private keys: %d files (0600)\n", privateKeyCount)
	}
	if publicKeyCount > 0 {
		fmt.Printf("• Public keys: %d files (0644/0600)\n", publicKeyCount)
	}

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("  ssh-add -l    # Check SSH agent\n")
	fmt.Printf("  ssh-add %s/id_rsa  # Add key to agent\n", opts.targetDir)

	return nil
}

// getLatestBackupName finds the most recent backup
func getLatestBackupName(provider interfaces.StorageProvider) (string, error) {
	ctx := context.Background()
	backups, err := provider.ListBackups(ctx)
	if err != nil {
		return "", err
	}

	if len(backups) == 0 {
		return "", fmt.Errorf("no backups found")
	}

	// For simplicity, assume backup names contain timestamps
	// In a real implementation, you'd parse metadata to find the latest
	var latest string
	for _, backup := range backups {
		if latest == "" || backup > latest {
			latest = backup
		}
	}

	return latest, nil
}

// parseVaultBackup converts Vault data back to backup structure
func parseVaultBackup(vaultData map[string]interface{}) (*ssh.BackupData, error) {
	backup := &ssh.BackupData{
		Files:    make(map[string]*ssh.FileData),
		Metadata: make(map[string]interface{}),
	}

	// Parse basic fields
	if version, ok := vaultData["version"].(string); ok {
		backup.Version = version
	}

	if hostname, ok := vaultData["hostname"].(string); ok {
		backup.Hostname = hostname
	}

	if username, ok := vaultData["username"].(string); ok {
		backup.Username = username
	}

	if sshDir, ok := vaultData["ssh_dir"].(string); ok {
		backup.SSHDir = sshDir
	}

	// Parse timestamp
	if timestampStr, ok := vaultData["timestamp"].(string); ok {
		if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			backup.Timestamp = timestamp
		}
	}

	// Parse metadata
	if metadata, ok := vaultData["metadata"].(map[string]interface{}); ok {
		backup.Metadata = metadata
	}

	// Parse files
	if filesData, ok := vaultData["files"].(map[string]interface{}); ok {
		for filename, fileDataInterface := range filesData {
			fileDataMap, ok := fileDataInterface.(map[string]interface{})
			if !ok {
				continue
			}

			fileData := &ssh.FileData{
				Filename: filename,
			}

			// Parse file metadata - handle both float64 and int64 from JSON
			// Note: stored permissions are just the permission bits (e.g., 420 for 0644)
			if perms, ok := fileDataMap["permissions"].(float64); ok {
				fileData.Permissions = os.FileMode(int(perms))
				log.Info().
					Str("file", filename).
					Float64("raw_perms", perms).
					Str("parsed_perms", fmt.Sprintf("%04o", fileData.Permissions&os.ModePerm)).
					Msg("Parsed permissions from float64")
			} else if perms, ok := fileDataMap["permissions"].(int64); ok {
				fileData.Permissions = os.FileMode(perms)
				log.Info().
					Str("file", filename).
					Int64("raw_perms", perms).
					Str("parsed_perms", fmt.Sprintf("%04o", fileData.Permissions&os.ModePerm)).
					Msg("Parsed permissions from int64")
			} else if perms, ok := fileDataMap["permissions"].(int); ok {
				fileData.Permissions = os.FileMode(perms)
				log.Info().
					Str("file", filename).
					Int("raw_perms", perms).
					Str("parsed_perms", fmt.Sprintf("%04o", fileData.Permissions&os.ModePerm)).
					Msg("Parsed permissions from int")
			} else {
				// If permissions are missing or invalid, use a safe fallback (0600) and log a warning
				log.Warn().
					Str("file", filename).
					Interface("permissions", fileDataMap["permissions"]).
					Msg("Missing or invalid permissions in backup data; using fallback permission 0600")
				fileData.Permissions = os.FileMode(0600)
			}

			if size, ok := fileDataMap["size"].(float64); ok {
				fileData.Size = int64(size)
			}

			if checksum, ok := fileDataMap["checksum"].(string); ok {
				fileData.Checksum = checksum
			}

			if modTimeStr, ok := fileDataMap["mod_time"].(string); ok {
				if modTime, err := time.Parse(time.RFC3339, modTimeStr); err == nil {
					fileData.ModTime = modTime
				}
			}

			// Parse file content
			if content, ok := fileDataMap["content"].(string); ok {
				fileData.Content = []byte(content)
				log.Debug().Str("file", filename).Msg("Parsed file content successfully")
			}

			// Parse key info if available
			if keyInfoData, ok := fileDataMap["key_info"].(map[string]interface{}); ok {
				keyInfo := &analyzer.KeyInfo{}

				if filename, ok := keyInfoData["filename"].(string); ok {
					keyInfo.Filename = filename
				}
				if keyType, ok := keyInfoData["type"].(string); ok {
					keyInfo.Type = analyzer.KeyType(keyType)
				}
				if format, ok := keyInfoData["format"].(string); ok {
					keyInfo.Format = analyzer.KeyFormat(format)
				}
				if size, ok := keyInfoData["size"].(float64); ok {
					keyInfo.Size = int64(size)
				}
				if perms, ok := keyInfoData["permissions"].(float64); ok {
					keyInfo.Permissions = os.FileMode(int(perms))
				} else if perms, ok := keyInfoData["permissions"].(int64); ok {
					keyInfo.Permissions = os.FileMode(perms)
				} else if perms, ok := keyInfoData["permissions"].(int); ok {
					keyInfo.Permissions = os.FileMode(perms)
				}
				if modTimeStr, ok := keyInfoData["mod_time"].(string); ok {
					if modTime, err := time.Parse(time.RFC3339, modTimeStr); err == nil {
						keyInfo.ModTime = modTime
					}
				}

				fileData.KeyInfo = keyInfo
				log.Debug().
					Str("file", filename).
					Str("key_type", string(keyInfo.Type)).
					Msg("Parsed key info successfully")
			}

			backup.Files[filename] = fileData
		}
	}

	return backup, nil
}

// displayRestoreSummary shows what will be restored
func displayRestoreSummary(backup *ssh.BackupData, targetDir string) {
	fmt.Printf("\n📥 Restore Summary\n")
	fmt.Printf("═════════════════\n")
	fmt.Printf("Backup from: %s (%s@%s)\n",
		backup.Timestamp.Format("2006-01-02 15:04:05"),
		backup.Username,
		backup.Hostname)
	fmt.Printf("Source SSH dir: %s\n", backup.SSHDir)
	fmt.Printf("Target dir: %s\n", targetDir)
	fmt.Printf("Files to restore: %d\n", len(backup.Files))

	// Show file list
	fmt.Printf("\n📄 Files:\n")
	for filename, fileData := range backup.Files {
		fmt.Printf("  • %s (%d bytes, %s)\n",
			filename,
			fileData.Size,
			fileData.Permissions.String())
	}
}

// interactiveRestoreSelection allows user to select files to restore
func interactiveRestoreSelection(backup *ssh.BackupData) error {
	fmt.Printf("\n🎯 Interactive File Selection\n")
	fmt.Printf("Select files to restore (y/n/a=all/q=quit):\n")

	for filename, fileData := range backup.Files {
		fmt.Printf("Restore '%s' [%d bytes, %s]? [y/N/a/q]: ",
			filename,
			fileData.Size,
			fileData.Permissions.String())

		var response string
		fmt.Scanln(&response)

		switch strings.ToLower(strings.TrimSpace(response)) {
		case "y", "yes":
			// Keep file
		case "a", "all":
			// Keep all remaining files
			return nil
		case "q", "quit":
			return fmt.Errorf("restore cancelled by user")
		default:
			// Remove file from restore
			delete(backup.Files, filename)
		}
	}

	return nil
}

// selectBackupInteractively shows available backups and lets user choose
func selectBackupInteractively(provider interfaces.StorageProvider) (string, error) {
	fmt.Printf("\n🔍 Finding available backups...\n")

	ctx := context.Background()

	// Get all backups
	backups, err := provider.ListBackups(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list backups: %w", err)
	}

	if len(backups) == 0 {
		return "", fmt.Errorf("no backups found")
	}

	// Get backup metadata for detailed display
	metadata, err := provider.GetMetadata(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get backup metadata")
	}

	// Prepare backup info with details
	var backupInfos []backupSelectionInfo
	for _, name := range backups {
		info := backupSelectionInfo{
			Name: name,
		}

		// Try to get metadata for this backup
		if metadata != nil {
			if backupsData, ok := metadata["backups"].(map[string]interface{}); ok {
				if backupData, ok := backupsData[name].(map[string]interface{}); ok {
					if timestampStr, ok := backupData["timestamp"].(string); ok {
						if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
							info.Timestamp = timestamp
						}
					}
					if fileCount, ok := backupData["file_count"].(float64); ok {
						info.FileCount = int(fileCount)
					}
					if totalSize, ok := backupData["total_size"].(float64); ok {
						info.TotalSize = int64(totalSize)
					}
					if hostname, ok := backupData["hostname"].(string); ok {
						info.Hostname = hostname
					}
					if username, ok := backupData["username"].(string); ok {
						info.Username = username
					}
				}
			}
		}

		backupInfos = append(backupInfos, info)
	}

	// Sort by timestamp (most recent first)
	sort.Slice(backupInfos, func(i, j int) bool {
		return backupInfos[i].Timestamp.After(backupInfos[j].Timestamp)
	})

	// Display available backups
	fmt.Printf("\n📦 Available Backups:\n")
	fmt.Printf("═══════════════════════\n\n")

	for i, info := range backupInfos {
		fmt.Printf("%d. 📦 %s\n", i+1, info.Name)

		if !info.Timestamp.IsZero() {
			fmt.Printf("   📅 Created: %s (%s ago)\n",
				info.Timestamp.Format("2006-01-02 15:04:05"),
				formatDuration(time.Since(info.Timestamp)))
		}

		if info.FileCount > 0 {
			fmt.Printf("   📄 Files: %d", info.FileCount)
			if info.TotalSize > 0 {
				fmt.Printf(", Size: %s", formatBytes(info.TotalSize))
			}
			fmt.Printf("\n")
		}

		if info.Hostname != "" || info.Username != "" {
			fmt.Printf("   💻 Source: %s@%s\n", info.Username, info.Hostname)
		}

		fmt.Printf("\n")
	}

	// Prompt for selection
	for {
		fmt.Printf("Select backup to restore [1-%d, q to quit]: ", len(backupInfos))

		var input string
		fmt.Scanln(&input)

		input = strings.TrimSpace(input)

		if input == "q" || input == "quit" {
			return "", fmt.Errorf("restore cancelled by user")
		}

		// Try to parse as number
		if choice := parseInt(input); choice >= 1 && choice <= len(backupInfos) {
			selected := backupInfos[choice-1]
			fmt.Printf("\n✅ Selected: %s\n", selected.Name)
			if !selected.Timestamp.IsZero() {
				fmt.Printf("   Created: %s\n", selected.Timestamp.Format("2006-01-02 15:04:05"))
			}
			if selected.FileCount > 0 {
				fmt.Printf("   Files: %d\n", selected.FileCount)
			}
			fmt.Printf("\n")
			return selected.Name, nil
		}

		fmt.Printf("Invalid selection. Please enter a number between 1 and %d, or 'q' to quit.\n", len(backupInfos))
	}
}

// backupSelectionInfo holds backup information for selection
type backupSelectionInfo struct {
	Name      string
	Timestamp time.Time
	FileCount int
	TotalSize int64
	Hostname  string
	Username  string
}

// parseInt safely parses an integer from string
func parseInt(s string) int {
	if num, err := fmt.Sscanf(s, "%d", new(int)); err == nil && num == 1 {
		var result int
		fmt.Sscanf(s, "%d", &result)
		return result
	}
	return 0
}
