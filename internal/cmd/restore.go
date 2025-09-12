package cmd

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-vault-keeper/internal/config"
	"github.com/rzago/ssh-vault-keeper/internal/ssh"
	"github.com/rzago/ssh-vault-keeper/internal/vault"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// newRestoreCommand creates the restore command
func newRestoreCommand(cfg *config.Config) *cobra.Command {
	var (
		backupName  string
		targetDir   string
		passphrase  string
		dryRun      bool
		overwrite   bool
		interactive bool
		fileFilter  []string
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
				backupName:  name,
				targetDir:   targetDir,
				passphrase:  passphrase,
				dryRun:      dryRun,
				overwrite:   overwrite,
				interactive: interactive,
				fileFilter:  fileFilter,
			})
		},
	}

	// Command-specific flags
	cmd.Flags().StringVar(&backupName, "backup", "", "Backup name to restore (default: most recent)")
	cmd.Flags().StringVar(&targetDir, "target-dir", cfg.Backup.SSHDir, "Target directory for restored files")
	cmd.Flags().StringVar(&passphrase, "passphrase", "", "Decryption passphrase (will prompt if not provided)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be restored without actually doing it")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing files without asking")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactively select files to restore")
	cmd.Flags().StringSliceVar(&fileFilter, "files", []string{}, "Only restore specific files (glob patterns)")

	return cmd
}

type restoreOptions struct {
	backupName  string
	targetDir   string
	passphrase  string
	dryRun      bool
	overwrite   bool
	interactive bool
	fileFilter  []string
}

func runRestore(cfg *config.Config, opts restoreOptions) error {
	log.Info().
		Str("backup_name", opts.backupName).
		Str("target_dir", opts.targetDir).
		Bool("dry_run", opts.dryRun).
		Msg("Starting restore process")

	// Connect to Vault
	fmt.Printf("Connecting to Vault...\n")
	vaultClient, err := vault.New(&cfg.Vault)
	if err != nil {
		return fmt.Errorf("failed to connect to Vault: %w", err)
	}
	defer vaultClient.Close()

	// Determine backup name
	backupName := opts.backupName
	if backupName == "" {
		backupName, err = getLatestBackupName(vaultClient)
		if err != nil {
			return fmt.Errorf("failed to find latest backup: %w", err)
		}
		fmt.Printf("Using most recent backup: %s\n", backupName)
	}

	// Retrieve backup from Vault
	fmt.Printf("Retrieving backup '%s' from Vault...\n", backupName)
	vaultData, err := vaultClient.GetBackup(backupName)
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

	// Get decryption passphrase
	passphrase := opts.passphrase
	if passphrase == "" {
		passphrase, err = promptRestorePassphrase("Enter decryption passphrase: ")
		if err != nil {
			return fmt.Errorf("failed to get passphrase: %w", err)
		}
	}

	// Initialize SSH handler
	sshHandler := ssh.New()

	// Decrypt backup
	fmt.Printf("Decrypting backup data...\n")
	if err := sshHandler.DecryptBackup(backupData, passphrase); err != nil {
		return fmt.Errorf("failed to decrypt backup: %w", err)
	}
	fmt.Printf("✓ Backup decrypted successfully\n")

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

	fmt.Printf("✓ Restore completed successfully\n")
	fmt.Printf("Files restored: %d\n", len(backupData.Files))

	// Remind user about permissions
	fmt.Printf("\n⚠️  Remember to check file permissions and SSH agent setup\n")
	fmt.Printf("Recommended next steps:\n")
	fmt.Printf("  chmod 700 %s\n", opts.targetDir)
	fmt.Printf("  ssh-add -l  # Check SSH agent\n")

	return nil
}

// getLatestBackupName finds the most recent backup
func getLatestBackupName(client *vault.Client) (string, error) {
	backups, err := client.ListBackups()
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

			// Parse file metadata
			if perms, ok := fileDataMap["permissions"].(float64); ok {
				fileData.Permissions = os.FileMode(int(perms))
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

			// Parse encrypted data
			if encryptedInterface, ok := fileDataMap["encrypted"]; ok {
				if _, ok := encryptedInterface.(map[string]interface{}); ok {
					// This would require more complex parsing of the crypto.EncryptedData structure
					// For now, create a basic structure
					log.Debug().Str("file", filename).Msg("Parsing encrypted data")
				}
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

// promptRestorePassphrase securely prompts for a passphrase
func promptRestorePassphrase(prompt string) (string, error) {
	fmt.Print(prompt)
	passphrase, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	
	if err != nil {
		return "", err
	}
	
	return string(passphrase), nil
}
