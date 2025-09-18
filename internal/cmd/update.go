package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-secret-keeper/internal/config"
	"github.com/rzago/ssh-secret-keeper/internal/update"
	"github.com/spf13/cobra"
)

// newUpdateCommand creates the update command
func newUpdateCommand(cfg *config.Config) *cobra.Command {
	var opts update.UpdateOptions
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update SSH Secret Keeper to the latest version",
		Long: `Check for and install updates to SSH Secret Keeper.

This command will:
1. Check for the latest release on GitHub
2. Download the appropriate binary for your platform
3. Verify the download integrity
4. Create a backup of the current binary
5. Replace the binary with the new version

Examples:
  # Check for updates without installing
  sshsk update --check

  # Update to the latest stable version
  sshsk update

  # Update to a specific version
  sshsk update --version v1.0.5

  # Include pre-release versions
  sshsk update --pre-release

  # Force reinstall even if already on latest
  sshsk update --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set up update configuration
			updateConfig := &update.UpdateConfig{
				GitHubRepo: "rafaelvzago/ssh-secret-keeper",
			}

			// Override with config if available
			if cfg != nil && cfg.Update != nil {
				updateConfig = &update.UpdateConfig{
					CheckOnStartup: cfg.Update.CheckOnStartup,
					AutoUpdate:     cfg.Update.AutoUpdate,
					Channel:        cfg.Update.Channel,
					CheckInterval:  cfg.Update.CheckInterval,
					GitHubRepo:     cfg.Update.GitHubRepo,
				}
			}

			// Create update service
			service := update.NewService(updateConfig)

			// Get current version
			currentVersion := Version
			if currentVersion == "" || currentVersion == "dev" {
				currentVersion = "v0.0.0" // Use a base version for dev builds
			}

			// Check only mode
			if checkOnly {
				return checkForUpdates(cmd, service, currentVersion, opts.PreRelease)
			}

			// Check for updates first
			status, err := service.CheckForUpdates(currentVersion, opts.PreRelease)
			if err != nil {
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			// Display update information
			displayUpdateStatus(cmd, status)

			if !status.UpdateAvailable && !opts.Force {
				cmd.Println("\n✅ You are already on the latest version!")
				return nil
			}

			// Ask for confirmation in interactive mode
			if !opts.Force && !confirmUpdate(cmd, status) {
				cmd.Println("Update cancelled.")
				return nil
			}

			// Perform the update
			cmd.Println("\n🚀 Starting update process...")

			// Set the options for check-only mode
			opts.CheckOnly = checkOnly

			if err := service.Update(currentVersion, opts); err != nil {
				return fmt.Errorf("update failed: %w", err)
			}

			cmd.Printf("\n✨ Successfully updated to %s!\n", status.LatestVersion)
			cmd.Println("Please restart any running sshsk processes to use the new version.")

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&opts.Version, "version", "", "Update to specific version")
	cmd.Flags().BoolVar(&checkOnly, "check", false, "Check for updates without installing")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force update even if already on latest")
	cmd.Flags().BoolVar(&opts.PreRelease, "pre-release", false, "Include pre-release versions")
	cmd.Flags().BoolVar(&opts.NoBackup, "no-backup", false, "Don't create backup of current binary")
	cmd.Flags().BoolVar(&opts.SkipChecksum, "skip-checksum", false, "Skip checksum verification (not recommended)")
	cmd.Flags().BoolVar(&opts.SkipVerify, "skip-verify", false, "Skip new binary verification")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	return cmd
}

// checkForUpdates checks for available updates and displays the result
func checkForUpdates(cmd *cobra.Command, service *update.Service, currentVersion string, includePreRelease bool) error {
	cmd.Println("🔍 Checking for updates...")

	status, err := service.CheckForUpdates(currentVersion, includePreRelease)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	displayUpdateStatus(cmd, status)

	if status.UpdateAvailable {
		cmd.Println("\n📦 An update is available!")
		cmd.Println("Run 'sshsk update' to install the latest version.")
	} else {
		cmd.Println("\n✅ You are on the latest version!")
	}

	return nil
}

// displayUpdateStatus displays the update status information
func displayUpdateStatus(cmd *cobra.Command, status *update.UpdateStatus) {
	cmd.Printf("\n📌 Current version: %s\n", formatVersion(status.CurrentVersion))
	cmd.Printf("🎯 Latest version:  %s", formatVersion(status.LatestVersion))

	if !status.PublishedAt.IsZero() {
		age := time.Since(status.PublishedAt)
		cmd.Printf(" (released %s)\n", formatUpdateDuration(age))
	} else {
		cmd.Println()
	}

	if status.UpdateAvailable {
		cmd.Println("🆕 Update available: Yes")

		// Show release notes if available
		if status.ReleaseNotes != "" {
			cmd.Println("\n📝 Release Notes:")
			displayReleaseNotes(cmd, status.ReleaseNotes)
		}
	} else {
		cmd.Println("🆕 Update available: No")
	}
}

// displayReleaseNotes formats and displays release notes
func displayReleaseNotes(cmd *cobra.Command, notes string) {
	lines := strings.Split(notes, "\n")
	maxLines := 10
	displayed := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Indent the line
		cmd.Printf("   %s\n", line)

		displayed++
		if displayed >= maxLines {
			remaining := len(lines) - displayed
			if remaining > 0 {
				cmd.Printf("   ... and %d more lines\n", remaining)
			}
			break
		}
	}
}

// confirmUpdate asks the user for confirmation
func confirmUpdate(cmd *cobra.Command, status *update.UpdateStatus) bool {
	// Check if --yes flag was provided
	if yes, _ := cmd.Flags().GetBool("yes"); yes {
		return true
	}

	// Check if we're in a non-interactive environment
	if !isInteractive() {
		log.Debug().Msg("Non-interactive environment detected, skipping confirmation")
		return true
	}

	cmd.Printf("\nDo you want to update to %s? (y/N): ", status.LatestVersion)

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// isInteractive checks if we're running in an interactive terminal
func isInteractive() bool {
	// Check if stdin is a terminal
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// formatVersion formats a version string for display
func formatVersion(version string) string {
	if version == "" {
		return "unknown"
	}
	if version == "dev" {
		return "dev (development build)"
	}
	return version
}

// formatUpdateDuration formats a duration for human-readable display
func formatUpdateDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	if days < 30 {
		return fmt.Sprintf("%d days ago", days)
	}
	if days < 365 {
		months := days / 30
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
	years := days / 365
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}
