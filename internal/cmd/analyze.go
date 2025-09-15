package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-vault-keeper/internal/analyzer"
	"github.com/rzago/ssh-vault-keeper/internal/config"
	"github.com/spf13/cobra"
)

// newAnalyzeCommand creates the analyze command
func newAnalyzeCommand(cfg *config.Config) *cobra.Command {
	var (
		sshDir     string
		outputJSON bool
		verbose    bool
	)

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze SSH directory and show detailed information",
		Long: `Analyze your SSH directory to understand the structure and types of keys.
This command helps you understand what will be backed up before running a backup.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalyze(cfg, analyzeOptions{
				sshDir:     sshDir,
				outputJSON: outputJSON,
				verbose:    verbose,
			})
		},
	}

	// Command-specific flags
	cmd.Flags().StringVar(&sshDir, "ssh-dir", cfg.Backup.SSHDir, "SSH directory to analyze")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output results in JSON format")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed information about each file")

	return cmd
}

type analyzeOptions struct {
	sshDir     string
	outputJSON bool
	verbose    bool
}

func runAnalyze(cfg *config.Config, opts analyzeOptions) error {
	log.Info().
		Str("ssh_dir", opts.sshDir).
		Bool("json_output", opts.outputJSON).
		Msg("Starting SSH directory analysis")

	// Initialize analyzer
	analyzer := analyzer.New()

	// Analyze directory
	result, err := analyzer.AnalyzeDirectory(opts.sshDir)
	if err != nil {
		return fmt.Errorf("failed to analyze SSH directory: %w", err)
	}

	// Output results
	if opts.outputJSON {
		return outputJSON(result)
	}

	return outputHuman(result, opts.verbose)
}

// outputJSON outputs the analysis results in JSON format
func outputJSON(result *analyzer.DetectionResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// outputHuman outputs the analysis results in human-readable format
func outputHuman(result *analyzer.DetectionResult, verbose bool) error {
	fmt.Printf("🔍 SSH Directory Analysis\n")
	fmt.Printf("════════════════════════\n\n")

	// Summary
	summary := result.Summary
	fmt.Printf("📊 Summary:\n")
	fmt.Printf("  Total files: %d\n", summary.TotalFiles)
	fmt.Printf("  Key pairs: %d\n", summary.KeyPairCount)
	fmt.Printf("  Service keys: %d\n", summary.ServiceKeys)
	fmt.Printf("  Personal keys: %d\n", summary.PersonalKeys)
	fmt.Printf("  Work keys: %d\n", summary.WorkKeys)
	fmt.Printf("  System files: %d\n", summary.SystemFiles)

	if summary.UnknownFiles > 0 {
		fmt.Printf("  ⚠️  Unknown files: %d\n", summary.UnknownFiles)
	}

	// Format breakdown
	if len(summary.FormatBreakdown) > 0 {
		fmt.Printf("\n📁 Format Breakdown:\n")
		for format, count := range summary.FormatBreakdown {
			fmt.Printf("  %s: %d\n", format, count)
		}
	}

	// Key pairs
	if len(result.KeyPairs) > 0 {
		fmt.Printf("\n🔑 Key Pairs:\n")
		for baseName, keyPair := range result.KeyPairs {
			status := ""
			if keyPair.PrivateKeyFile != "" && keyPair.PublicKeyFile != "" {
				status = "✓ Complete pair"
			} else if keyPair.PrivateKeyFile != "" {
				status = "⚠️  Private key only"
			} else {
				status = "⚠️  Public key only"
			}

			fmt.Printf("  • %s - %s\n", baseName, status)
			if verbose {
				if keyPair.PrivateKeyFile != "" {
					fmt.Printf("    Private: %s\n", keyPair.PrivateKeyFile)
				}
				if keyPair.PublicKeyFile != "" {
					fmt.Printf("    Public: %s\n", keyPair.PublicKeyFile)
				}
			}
		}
	}

	// Categories
	if len(result.Categories) > 0 {
		fmt.Printf("\n📂 Categories:\n")
		for category, files := range result.Categories {
			fmt.Printf("  %s (%d files):\n", category, len(files))
			for _, file := range files {
				icon := getFileIcon(file.Type)
				fmt.Printf("    %s %s", icon, file.Filename)
				if file.Service != "" {
					fmt.Printf(" [%s]", file.Service)
				}
				fmt.Printf("\n")

				if verbose {
					fmt.Printf("      Type: %s, Format: %s, Size: %d bytes\n",
						file.Type, file.Format, file.Size)
					fmt.Printf("      Permissions: %s, Modified: %s\n",
						file.Permissions.String(),
						file.ModTime.Format("2006-01-02 15:04"))
				}
			}
		}
	}

	// System files
	if len(result.SystemFiles) > 0 {
		fmt.Printf("\n⚙️  System Files:\n")
		for _, file := range result.SystemFiles {
			icon := getFileIcon(file.Type)
			fmt.Printf("  %s %s (%d bytes)\n", icon, file.Filename, file.Size)
		}
	}

	// Unknown files
	if len(result.UnknownFiles) > 0 {
		fmt.Printf("\n❓ Unknown Files:\n")
		for _, file := range result.UnknownFiles {
			fmt.Printf("  • %s (%d bytes)\n", file.Filename, file.Size)
			if verbose {
				fmt.Printf("    Permissions: %s, Modified: %s\n",
					file.Permissions.String(),
					file.ModTime.Format("2006-01-02 15:04"))
			}
		}
	}

	// Recommendations
	fmt.Printf("\n💡 Recommendations:\n")
	if summary.UnknownFiles > 0 {
		fmt.Printf("  • Review unknown files - they won't be backed up automatically\n")
	}
	if summary.KeyPairCount == 0 {
		fmt.Printf("  • No SSH key pairs found - consider generating some with ssh-keygen\n")
	}

	// Check for incomplete pairs
	incompletePairs := 0
	for _, keyPair := range result.KeyPairs {
		if keyPair.PrivateKeyFile == "" || keyPair.PublicKeyFile == "" {
			incompletePairs++
		}
	}
	if incompletePairs > 0 {
		fmt.Printf("  • %d incomplete key pairs found - consider generating missing parts\n", incompletePairs)
	}

	fmt.Printf("  • Run 'ssh-vault-keeper backup' to create a secure backup\n")

	return nil
}

// getFileIcon returns an appropriate icon for the file type
func getFileIcon(fileType analyzer.KeyType) string {
	switch fileType {
	case analyzer.KeyTypePrivate:
		return "🔐"
	case analyzer.KeyTypePublic:
		return "🔑"
	case analyzer.KeyTypeConfig:
		return "⚙️"
	case analyzer.KeyTypeHosts:
		return "🌐"
	case analyzer.KeyTypeAuthorized:
		return "🎫"
	case analyzer.KeyTypeCertificate:
		return "📜"
	default:
		return "📄"
	}
}
