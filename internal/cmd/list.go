package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-secret-keeper/internal/config"
	"github.com/rzago/ssh-secret-keeper/internal/interfaces"
	"github.com/rzago/ssh-secret-keeper/internal/storage"
	"github.com/rzago/ssh-secret-keeper/internal/vault"
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

	// Create storage provider via factory
	factory := storage.NewFactory()
	storageProvider, err := factory.CreateStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to create storage provider: %w", err)
	}
	defer storageProvider.Close()

	ctx := context.Background()

	// List backups
	backupNames, err := storageProvider.ListBackups(ctx)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if len(backupNames) == 0 {
		// Enhanced: Check for backups in other storage strategies
		if err := handleNoBackupsFound(cfg, storageProvider, ctx, opts.outputJSON); err != nil {
			log.Debug().Err(err).Msg("Error checking for cross-machine backups")
		}
		return nil
	}

	// Get detailed information if requested
	var backups []backupInfo
	if opts.detailed {
		metadata, err := storageProvider.GetMetadata(ctx)
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
	fmt.Printf("  • sshsk restore <backup-name>  - Restore a specific backup\n")
	fmt.Printf("  • sshsk restore               - Restore most recent backup\n")

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

// AlternativeBackupLocation represents backups found in alternative storage paths
type AlternativeBackupLocation struct {
	Strategy      vault.StorageStrategy
	Path          string
	Description   string
	BackupCount   int
	SampleBackups []string
}

// handleNoBackupsFound checks for backups in other storage strategies and provides guidance
func handleNoBackupsFound(cfg *config.Config, storageProvider interfaces.StorageProvider, ctx context.Context, outputJSON bool) error {
	if outputJSON {
		// For JSON output, just return empty result
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(map[string]interface{}{
			"backups": []backupInfo{},
			"count":   0,
		})
	}

	fmt.Printf("No backups found in %s storage.\n", storageProvider.GetProviderType())

	// Only check for cross-machine backups if using Vault storage
	if storageProvider.GetProviderType() != "vault" {
		fmt.Printf("Use 'sshsk backup' to create your first backup.\n")
		return nil
	}

	// Check current strategy
	currentStrategy, err := vault.ParseStrategy(cfg.Vault.StorageStrategy)
	if err != nil {
		fmt.Printf("Use 'sshsk backup' to create your first backup.\n")
		return nil
	}

	// If already using universal, no need to check further
	if currentStrategy == vault.StrategyUniversal {
		fmt.Printf("Use 'sshsk backup' to create your first backup.\n")
		return nil
	}

	// Check for backups in alternative storage strategies
	alternativeBackups := findAlternativeBackups(cfg, ctx)

	if len(alternativeBackups) > 0 {
		fmt.Printf("\n🔍 **Cross-Machine Backup Detection**\n")
		fmt.Printf("═════════════════════════════════════\n")
		fmt.Printf("Found %d backup location(s) with existing backups:\n\n", len(alternativeBackups))

		for _, alt := range alternativeBackups {
			fmt.Printf("📦 %s (%d backups)\n", alt.Description, alt.BackupCount)
			fmt.Printf("   Path: %s\n", alt.Path)
			if len(alt.SampleBackups) > 0 {
				sampleCount := min(3, len(alt.SampleBackups))
				fmt.Printf("   Recent: %s\n", strings.Join(alt.SampleBackups[:sampleCount], ", "))
			}
			fmt.Printf("\n")
		}

		fmt.Printf("💡 **Solution - Enable Cross-Machine Restore:**\n")
		fmt.Printf("These backups are likely from a different machine or user.\n")
		fmt.Printf("To access them from this machine, migrate to universal storage:\n\n")

		// Determine the most likely source strategy
		mostLikelySource := determineMostLikelySource(alternativeBackups)
		if mostLikelySource != "" {
			fmt.Printf("🚀 **Quick Fix:**\n")
			fmt.Printf("   sshsk migrate --from %s --to universal --dry-run\n", mostLikelySource)
			fmt.Printf("   sshsk migrate --from %s --to universal --cleanup\n\n", mostLikelySource)
		}

		fmt.Printf("📚 **Learn More:**\n")
		fmt.Printf("   sshsk migrate-status  # Show all available strategies\n")
		fmt.Printf("   sshsk migrate --help  # Migration command help\n")
	} else {
		fmt.Printf("Use 'sshsk backup' to create your first backup.\n")
	}

	return nil
}

// findAlternativeBackups scans for backups in other storage strategies
func findAlternativeBackups(cfg *config.Config, ctx context.Context) []AlternativeBackupLocation {
	var alternatives []AlternativeBackupLocation

	// Check current strategy
	currentStrategy, err := vault.ParseStrategy(cfg.Vault.StorageStrategy)
	if err != nil {
		return alternatives
	}

	// Check each strategy except the current one
	strategies := []vault.StorageStrategy{
		vault.StrategyMachineUser, // Most likely for cross-machine issues
		vault.StrategyUser,
		vault.StrategyUniversal,
	}

	for _, strategy := range strategies {
		if strategy == currentStrategy {
			continue
		}

		if alt := scanStrategyForBackups(cfg, strategy, ctx); alt != nil {
			alternatives = append(alternatives, *alt)
		}
	}

	return alternatives
}

// scanStrategyForBackups scans a specific strategy for backups
func scanStrategyForBackups(cfg *config.Config, strategy vault.StorageStrategy, ctx context.Context) *AlternativeBackupLocation {
	// Create a temporary path generator for this strategy
	pathGenerator := vault.NewPathGenerator(strategy, cfg.Vault.CustomPrefix, cfg.Vault.BackupNamespace)
	basePath, err := pathGenerator.GenerateBasePath()
	if err != nil {
		return nil
	}

	// Try to create a migration service to list backups in this path
	migrationService, err := vault.NewMigrationService(&cfg.Vault, strategy, vault.StrategyUniversal)
	if err != nil {
		return nil
	}

	// List backups in this alternative location
	backups, err := migrationService.ListBackupsToMigrate(ctx)
	if err != nil || len(backups) == 0 {
		return nil
	}

	// Sort backups to get most recent ones first
	sort.Strings(backups)
	if len(backups) > 1 {
		// Reverse to get most recent first (assuming timestamp-based naming)
		for i, j := 0, len(backups)-1; i < j; i, j = i+1, j-1 {
			backups[i], backups[j] = backups[j], backups[i]
		}
	}

	return &AlternativeBackupLocation{
		Strategy:      strategy,
		Path:          basePath,
		Description:   pathGenerator.GetStrategyDescription(),
		BackupCount:   len(backups),
		SampleBackups: backups,
	}
}

// determineMostLikelySource determines the most likely source strategy for migration
func determineMostLikelySource(alternatives []AlternativeBackupLocation) string {
	if len(alternatives) == 0 {
		return ""
	}

	// Prefer machine-user strategy as it's the most common legacy case
	for _, alt := range alternatives {
		if alt.Strategy == vault.StrategyMachineUser {
			return string(alt.Strategy)
		}
	}

	// Otherwise, return the one with the most backups
	var best *AlternativeBackupLocation
	for i := range alternatives {
		if best == nil || alternatives[i].BackupCount > best.BackupCount {
			best = &alternatives[i]
		}
	}

	if best != nil {
		return string(best.Strategy)
	}

	return ""
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
