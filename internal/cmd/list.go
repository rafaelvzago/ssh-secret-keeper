package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-vault-keeper/internal/config"
	"github.com/rzago/ssh-vault-keeper/internal/vault"
	"github.com/spf13/cobra"
)

// newListCommand creates the list command
func newListCommand(cfg *config.Config) *cobra.Command {
	var (
		outputJSON bool
		detailed   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available backups in Vault",
		Long: `List all SSH backups stored in Vault with their metadata.
Shows backup names, timestamps, file counts, and other useful information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cfg, listOptions{
				outputJSON: outputJSON,
				detailed:   detailed,
			})
		},
	}

	// Command-specific flags
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output results in JSON format")
	cmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed information about each backup")

	return cmd
}

type listOptions struct {
	outputJSON bool
	detailed   bool
}

type backupInfo struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	FileCount int       `json:"file_count"`
	TotalSize int64     `json:"total_size"`
	Hostname  string    `json:"hostname"`
	Username  string    `json:"username"`
}

func runList(cfg *config.Config, opts listOptions) error {
	log.Info().
		Bool("json_output", opts.outputJSON).
		Bool("detailed", opts.detailed).
		Msg("Listing backups")

	// Connect to Vault
	vaultClient, err := vault.New(&cfg.Vault)
	if err != nil {
		return fmt.Errorf("failed to connect to Vault: %w", err)
	}
	defer vaultClient.Close()

	// List backups
	backupNames, err := vaultClient.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if len(backupNames) == 0 {
		fmt.Printf("No backups found in Vault.\n")
		fmt.Printf("Use 'ssh-vault-keeper backup' to create your first backup.\n")
		return nil
	}

	// Get detailed information if requested
	var backups []backupInfo
	if opts.detailed {
		metadata, err := vaultClient.GetMetadata()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get metadata, using basic info")
		}

		for _, name := range backupNames {
			backup := backupInfo{Name: name}

			// Try to get metadata for this backup
			if metadata != nil {
				if backupsData, ok := metadata["backups"].(map[string]interface{}); ok {
					if backupData, ok := backupsData[name].(map[string]interface{}); ok {
						if timestampStr, ok := backupData["timestamp"].(string); ok {
							if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
								backup.Timestamp = timestamp
							}
						}
						if fileCount, ok := backupData["file_count"].(float64); ok {
							backup.FileCount = int(fileCount)
						}
						if totalSize, ok := backupData["total_size"].(float64); ok {
							backup.TotalSize = int64(totalSize)
						}
						if hostname, ok := backupData["hostname"].(string); ok {
							backup.Hostname = hostname
						}
						if username, ok := backupData["username"].(string); ok {
							backup.Username = username
						}
					}
				}
			}

			backups = append(backups, backup)
		}

		// Sort by timestamp (most recent first)
		sort.Slice(backups, func(i, j int) bool {
			return backups[i].Timestamp.After(backups[j].Timestamp)
		})
	} else {
		// Basic listing
		for _, name := range backupNames {
			backups = append(backups, backupInfo{Name: name})
		}
		// Sort alphabetically for basic listing
		sort.Slice(backups, func(i, j int) bool {
			return backups[i].Name < backups[j].Name
		})
	}

	// Output results
	if opts.outputJSON {
		return outputJSONList(backups)
	}

	return outputHumanList(backups, opts.detailed)
}

// outputJSONList outputs the backup list in JSON format
func outputJSONList(backups []backupInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(map[string]interface{}{
		"backups": backups,
		"count":   len(backups),
	})
}

// outputHumanList outputs the backup list in human-readable format
func outputHumanList(backups []backupInfo, detailed bool) error {
	fmt.Printf("📦 Available Backups\n")
	fmt.Printf("═══════════════════\n\n")

	if !detailed {
		fmt.Printf("Found %d backup(s):\n\n", len(backups))
		for i, backup := range backups {
			fmt.Printf("%d. %s\n", i+1, backup.Name)
		}
		fmt.Printf("\nUse --detailed for more information about each backup.\n")
		return nil
	}

	// Detailed output
	fmt.Printf("Found %d backup(s) (most recent first):\n\n", len(backups))

	for i, backup := range backups {
		fmt.Printf("%d. 📦 %s\n", i+1, backup.Name)

		if !backup.Timestamp.IsZero() {
			fmt.Printf("   📅 Created: %s (%s ago)\n",
				backup.Timestamp.Format("2006-01-02 15:04:05"),
				formatDuration(time.Since(backup.Timestamp)))
		}

		if backup.FileCount > 0 {
			fmt.Printf("   📄 Files: %d\n", backup.FileCount)
		}

		if backup.TotalSize > 0 {
			fmt.Printf("   💾 Size: %s\n", formatBytes(backup.TotalSize))
		}

		if backup.Hostname != "" || backup.Username != "" {
			fmt.Printf("   💻 Source: %s@%s\n", backup.Username, backup.Hostname)
		}

		if i < len(backups)-1 {
			fmt.Printf("\n")
		}
	}

	fmt.Printf("\n💡 Commands:\n")
	fmt.Printf("  • ssh-vault-keeper restore <backup-name>  - Restore a specific backup\n")
	fmt.Printf("  • ssh-vault-keeper restore               - Restore most recent backup\n")

	return nil
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}

// formatBytes formats bytes in a human-readable way
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
