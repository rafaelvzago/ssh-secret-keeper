package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-secret-keeper/internal/config"
	"github.com/rzago/ssh-secret-keeper/internal/vault"
	"github.com/spf13/cobra"
)

// newMigrateCommand creates the migrate command
func newMigrateCommand(cfg *config.Config) *cobra.Command {
	var (
		fromStrategy string
		toStrategy   string
		dryRun       bool
		cleanup      bool
		force        bool
	)

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate backups between storage strategies",
		Long: `Migrate backups between different storage strategies.

This command allows you to move existing backups from one storage strategy to another,
enabling cross-machine and cross-user restore capabilities.

Available strategies:
  - universal:    Shared storage across machines and users (recommended)
  - user:         User-scoped storage, shared across machines
  - machine-user: Machine-user scoped storage (legacy)
  - custom:       Custom storage prefix

Examples:
  # Migrate from machine-user to universal storage (dry run)
  sshsk migrate --from machine-user --to universal --dry-run

  # Migrate and cleanup source backups
  sshsk migrate --from machine-user --to universal --cleanup

  # Show what would be migrated without doing it
  sshsk migrate --from machine-user --to universal --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrate(cfg, migrateOptions{
				fromStrategy: fromStrategy,
				toStrategy:   toStrategy,
				dryRun:       dryRun,
				cleanup:      cleanup,
				force:        force,
			})
		},
	}

	// Command-specific flags
	cmd.Flags().StringVar(&fromStrategy, "from", "", "Source storage strategy (required)")
	cmd.Flags().StringVar(&toStrategy, "to", "", "Destination storage strategy (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be migrated without actually doing it")
	cmd.Flags().BoolVar(&cleanup, "cleanup", false, "Remove source backups after successful migration")
	cmd.Flags().BoolVar(&force, "force", false, "Skip validation and confirmation prompts")

	// Mark required flags
	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")

	return cmd
}

type migrateOptions struct {
	fromStrategy string
	toStrategy   string
	dryRun       bool
	cleanup      bool
	force        bool
}

func runMigrate(cfg *config.Config, opts migrateOptions) error {
	log.Info().
		Str("from_strategy", opts.fromStrategy).
		Str("to_strategy", opts.toStrategy).
		Bool("dry_run", opts.dryRun).
		Bool("cleanup", opts.cleanup).
		Msg("Starting backup migration")

	// Parse and validate strategies
	fromStrategy, err := vault.ParseStrategy(opts.fromStrategy)
	if err != nil {
		return fmt.Errorf("invalid source strategy: %w", err)
	}

	toStrategy, err := vault.ParseStrategy(opts.toStrategy)
	if err != nil {
		return fmt.Errorf("invalid destination strategy: %w", err)
	}

	// Create migration service
	migrationService, err := vault.NewMigrationService(&cfg.Vault, fromStrategy, toStrategy)
	if err != nil {
		return fmt.Errorf("failed to create migration service: %w", err)
	}

	ctx := context.Background()

	// Validate migration
	fmt.Printf("Validating migration from %s to %s...\n", fromStrategy, toStrategy)
	validation, err := migrationService.ValidateMigration(ctx)
	if err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	if !validation.Valid {
		fmt.Printf("❌ Migration validation failed:\n")
		for _, errMsg := range validation.Errors {
			fmt.Printf("  • %s\n", errMsg)
		}
		return fmt.Errorf("migration cannot proceed due to validation errors")
	}

	// Display validation results
	fmt.Printf("✅ Migration validation passed\n")
	fmt.Printf("Source backups found: %d\n", validation.SourceBackupCount)

	if len(validation.Benefits) > 0 {
		fmt.Printf("\n📈 Benefits:\n")
		for _, benefit := range validation.Benefits {
			fmt.Printf("  • %s\n", benefit)
		}
	}

	if len(validation.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings:\n")
		for _, warning := range validation.Warnings {
			fmt.Printf("  • %s\n", warning)
		}
	}

	// Get user confirmation unless forced
	if !opts.force && !opts.dryRun {
		fmt.Printf("\nProceed with migration? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Printf("Migration cancelled\n")
			return nil
		}
	}

	// Perform migration
	fmt.Printf("\n🚀 Starting migration...\n")
	result, err := migrationService.MigrateAllBackups(ctx, opts.dryRun)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Display results
	fmt.Printf("\n%s", result.GetMigrationSummary())

	if len(result.FailedBackups) > 0 {
		fmt.Printf("❌ Some backups failed to migrate. Check logs for details.\n")
		return fmt.Errorf("migration completed with %d failures", len(result.FailedBackups))
	}

	if opts.dryRun {
		fmt.Printf("✅ Dry run completed successfully. Use --dry-run=false to perform actual migration.\n")
		return nil
	}

	fmt.Printf("✅ Migration completed successfully!\n")

	// Cleanup source backups if requested
	if opts.cleanup && len(result.MigratedBackups) > 0 {
		fmt.Printf("\n🧹 Cleaning up source backups...\n")

		if !opts.force {
			fmt.Printf("Delete %d source backups? [y/N]: ", len(result.MigratedBackups))
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Printf("Cleanup cancelled - source backups preserved\n")
				return nil
			}
		}

		if err := migrationService.CleanupSourceBackups(ctx, result.MigratedBackups, false); err != nil {
			log.Error().Err(err).Msg("Failed to cleanup some source backups")
			fmt.Printf("⚠️  Some source backups could not be deleted. Check logs for details.\n")
		} else {
			fmt.Printf("✅ Source backups cleaned up successfully\n")
		}
	}

	return nil
}

// newMigrateStatusCommand creates a command to show migration status and available strategies
func newMigrateStatusCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate-status",
		Short: "Show available storage strategies and migration options",
		Long: `Display information about available storage strategies and potential migrations.

This command helps you understand the current storage configuration and
available migration options.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrateStatus(cfg)
		},
	}

	return cmd
}

func runMigrateStatus(cfg *config.Config) error {
	fmt.Printf("SSH Secret Keeper - Storage Strategy Information\n")
	fmt.Printf("==============================================\n\n")

	// Show current configuration
	fmt.Printf("Current Configuration:\n")
	fmt.Printf("  Storage Strategy: %s\n", cfg.Vault.StorageStrategy)
	if cfg.Vault.BackupNamespace != "" {
		fmt.Printf("  Backup Namespace: %s\n", cfg.Vault.BackupNamespace)
	}
	if cfg.Vault.CustomPrefix != "" {
		fmt.Printf("  Custom Prefix: %s\n", cfg.Vault.CustomPrefix)
	}

	// Generate current path
	currentStrategy, err := vault.ParseStrategy(cfg.Vault.StorageStrategy)
	if err != nil {
		fmt.Printf("  ⚠️  Invalid current strategy: %v\n", err)
	} else {
		pathGenerator := vault.NewPathGenerator(currentStrategy, cfg.Vault.CustomPrefix, cfg.Vault.BackupNamespace)
		currentPath, err := pathGenerator.GenerateBasePath()
		if err != nil {
			fmt.Printf("  ⚠️  Cannot generate current path: %v\n", err)
		} else {
			fmt.Printf("  Current Storage Path: %s\n", currentPath)
			fmt.Printf("  Description: %s\n", pathGenerator.GetStrategyDescription())
		}
	}

	fmt.Printf("\nAvailable Storage Strategies:\n")
	strategies := vault.GetAllStrategies()
	for strategy, description := range strategies {
		fmt.Printf("  %s: %s\n", strategy, description)
	}

	fmt.Printf("\nMigration Examples:\n")
	fmt.Printf("  # Migrate from legacy machine-user to universal storage\n")
	fmt.Printf("  sshsk migrate --from machine-user --to universal --dry-run\n\n")
	fmt.Printf("  # Migrate to user-scoped storage for shared Vault\n")
	fmt.Printf("  sshsk migrate --from machine-user --to user --cleanup\n\n")
	fmt.Printf("  # Use custom prefix for team organization\n")
	fmt.Printf("  sshsk migrate --from universal --to custom --dry-run\n")
	fmt.Printf("  # (Note: Set custom_prefix in config for custom strategy)\n\n")

	fmt.Printf("Cross-Machine Restore:\n")
	fmt.Printf("  ✅ universal: Backups accessible from any machine/user\n")
	fmt.Printf("  ✅ user: Backups accessible from any machine for same user\n")
	fmt.Printf("  ❌ machine-user: Backups tied to specific machine-user combination\n")
	fmt.Printf("  ✅ custom: Depends on prefix configuration\n")

	return nil
}
