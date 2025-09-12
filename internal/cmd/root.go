package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-vault-keeper/internal/config"
	"github.com/spf13/cobra"
)

// Version information
var (
	Version   = "1.0.0"
	BuildTime = "unknown"
	GitHash   = "unknown"
)

// NewRootCommand creates the root command
func NewRootCommand(cfg *config.Config) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "ssh-vault-keeper",
		Short: "Securely backup SSH keys to HashiCorp Vault",
		Long: `SSH Vault Keeper is a tool for securely backing up SSH keys and configuration
to HashiCorp Vault with client-side encryption. It provides intelligent analysis
of SSH directories and flexible backup/restore operations.`,
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// This runs before every command
			log.Debug().
				Str("command", cmd.Name()).
				Strs("args", args).
				Msg("Executing command")
		},
	}

	// Global flags
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging")
	rootCmd.PersistentFlags().Bool("quiet", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().String("config", "", "Configuration file path")

	// Add subcommands
	rootCmd.AddCommand(
		newInitCommand(cfg),
		newBackupCommand(cfg),
		newRestoreCommand(cfg),
		newListCommand(cfg),
		newAnalyzeCommand(cfg),
		newStatusCommand(cfg),
		newVersionCommand(),
	)

	return rootCmd
}

// newVersionCommand creates the version command
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("SSH Vault Keeper %s\n", Version)
			fmt.Printf("Built at: %s\n", BuildTime)
			fmt.Printf("Git hash: %s\n", GitHash)
		},
	}
}

// setupLogging configures logging based on flags
func setupLogging(cmd *cobra.Command) {
	verbose, _ := cmd.Flags().GetBool("verbose")
	quiet, _ := cmd.Flags().GetBool("quiet")

	if verbose {
		log.Debug().Msg("Verbose logging enabled")
	} else if quiet {
		// Only show errors when quiet
		log.Info().Msg("Quiet mode enabled")
	}
}

// handleError handles common errors and exits with appropriate code
func handleError(err error, message string) {
	if err != nil {
		log.Error().Err(err).Msg(message)
		os.Exit(1)
	}
}
