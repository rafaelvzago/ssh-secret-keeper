package vault

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/rzago/ssh-secret-keeper/internal/config"
)

// MigrationService handles migration of backups between different storage strategies
type MigrationService struct {
	client       *api.Client
	mountPath    string
	fromPath     string
	toPath       string
	fromStrategy StorageStrategy
	toStrategy   StorageStrategy
}

// NewMigrationService creates a new migration service
func NewMigrationService(cfg *config.VaultConfig, fromStrategy, toStrategy StorageStrategy) (*MigrationService, error) {
	client, err := createVaultClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	// Generate paths for source and destination strategies
	fromGenerator := NewPathGenerator(fromStrategy, cfg.CustomPrefix, cfg.BackupNamespace)
	fromPath, err := fromGenerator.GenerateBasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to generate source path: %w", err)
	}

	toGenerator := NewPathGenerator(toStrategy, cfg.CustomPrefix, cfg.BackupNamespace)
	toPath, err := toGenerator.GenerateBasePath()
	if err != nil {
		return nil, fmt.Errorf("failed to generate destination path: %w", err)
	}

	return &MigrationService{
		client:       client,
		mountPath:    cfg.MountPath,
		fromPath:     fromPath,
		toPath:       toPath,
		fromStrategy: fromStrategy,
		toStrategy:   toStrategy,
	}, nil
}

// ListBackupsToMigrate lists all backups in the source location
func (m *MigrationService) ListBackupsToMigrate(ctx context.Context) ([]string, error) {
	sourcePath := fmt.Sprintf("%s/metadata/%s/backups", m.mountPath, m.fromPath)

	secret, err := m.client.Logical().ListWithContext(ctx, sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list source backups: %w", err)
	}

	if secret == nil {
		return []string{}, nil
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	var backups []string
	for _, key := range keys {
		if keyStr, ok := key.(string); ok {
			backups = append(backups, keyStr)
		}
	}

	return backups, nil
}

// MigrateBackup migrates a single backup from source to destination
func (m *MigrationService) MigrateBackup(ctx context.Context, backupName string) error {
	log.Info().
		Str("backup", backupName).
		Str("from_path", m.fromPath).
		Str("to_path", m.toPath).
		Msg("Starting backup migration")

	// Read backup from source location
	sourcePath := fmt.Sprintf("%s/data/%s/backups/%s", m.mountPath, m.fromPath, backupName)
	secret, err := m.client.Logical().ReadWithContext(ctx, sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source backup %s: %w", backupName, err)
	}

	if secret == nil {
		return fmt.Errorf("backup %s not found at source location", backupName)
	}

	// Extract data from KV v2 format
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid backup data format for %s", backupName)
	}

	// Add migration metadata
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		metadata["migrated_from"] = m.fromPath
		metadata["migrated_to"] = m.toPath
		metadata["migrated_at"] = time.Now().Format(time.RFC3339)
		metadata["migration_strategy"] = fmt.Sprintf("%s->%s", m.fromStrategy, m.toStrategy)
	}

	// Write backup to destination location
	destPath := fmt.Sprintf("%s/data/%s/backups/%s", m.mountPath, m.toPath, backupName)
	wrappedData := map[string]interface{}{
		"data": data,
		"metadata": map[string]interface{}{
			"migrated_at": time.Now().Format(time.RFC3339),
			"source_path": sourcePath,
			"dest_path":   destPath,
		},
	}

	_, err = m.client.Logical().WriteWithContext(ctx, destPath, wrappedData)
	if err != nil {
		return fmt.Errorf("failed to write backup %s to destination: %w", backupName, err)
	}

	log.Info().
		Str("backup", backupName).
		Str("source", sourcePath).
		Str("destination", destPath).
		Msg("Backup migrated successfully")

	return nil
}

// MigrateAllBackups migrates all backups from source to destination
func (m *MigrationService) MigrateAllBackups(ctx context.Context, dryRun bool) (*MigrationResult, error) {
	backups, err := m.ListBackupsToMigrate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups for migration: %w", err)
	}

	result := &MigrationResult{
		FromStrategy:    m.fromStrategy,
		ToStrategy:      m.toStrategy,
		FromPath:        m.fromPath,
		ToPath:          m.toPath,
		TotalBackups:    len(backups),
		MigratedBackups: []string{},
		FailedBackups:   []string{},
		DryRun:          dryRun,
		StartTime:       time.Now(),
	}

	if len(backups) == 0 {
		log.Info().
			Str("from_path", m.fromPath).
			Msg("No backups found to migrate")
		result.EndTime = time.Now()
		return result, nil
	}

	log.Info().
		Int("backup_count", len(backups)).
		Str("from_strategy", string(m.fromStrategy)).
		Str("to_strategy", string(m.toStrategy)).
		Bool("dry_run", dryRun).
		Msg("Starting migration of all backups")

	for _, backupName := range backups {
		if dryRun {
			log.Info().
				Str("backup", backupName).
				Msg("[DRY RUN] Would migrate backup")
			result.MigratedBackups = append(result.MigratedBackups, backupName)
			continue
		}

		if err := m.MigrateBackup(ctx, backupName); err != nil {
			log.Error().
				Err(err).
				Str("backup", backupName).
				Msg("Failed to migrate backup")
			result.FailedBackups = append(result.FailedBackups, backupName)
			continue
		}

		result.MigratedBackups = append(result.MigratedBackups, backupName)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	log.Info().
		Int("migrated", len(result.MigratedBackups)).
		Int("failed", len(result.FailedBackups)).
		Dur("duration", result.Duration).
		Msg("Migration completed")

	return result, nil
}

// CleanupSourceBackups removes backups from the source location after successful migration
func (m *MigrationService) CleanupSourceBackups(ctx context.Context, backupNames []string, dryRun bool) error {
	if len(backupNames) == 0 {
		return nil
	}

	log.Info().
		Int("backup_count", len(backupNames)).
		Bool("dry_run", dryRun).
		Msg("Starting cleanup of source backups")

	for _, backupName := range backupNames {
		sourcePath := fmt.Sprintf("%s/data/%s/backups/%s", m.mountPath, m.fromPath, backupName)

		if dryRun {
			log.Info().
				Str("backup", backupName).
				Str("path", sourcePath).
				Msg("[DRY RUN] Would delete source backup")
			continue
		}

		_, err := m.client.Logical().DeleteWithContext(ctx, sourcePath)
		if err != nil {
			log.Error().
				Err(err).
				Str("backup", backupName).
				Str("path", sourcePath).
				Msg("Failed to delete source backup")
			continue
		}

		log.Info().
			Str("backup", backupName).
			Str("path", sourcePath).
			Msg("Source backup deleted successfully")
	}

	return nil
}

// ValidateMigration validates that the migration is safe and feasible
func (m *MigrationService) ValidateMigration(ctx context.Context) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Warnings: []string{},
		Errors:   []string{},
	}

	// Check if source and destination are the same
	if m.fromPath == m.toPath {
		result.Valid = false
		result.Errors = append(result.Errors, "Source and destination paths are identical - no migration needed")
		return result, nil
	}

	// Check if source location exists and has backups
	sourceBackups, err := m.ListBackupsToMigrate(ctx)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Cannot access source location: %v", err))
		return result, nil
	}

	if len(sourceBackups) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "No backups found at source location")
		return result, nil
	}

	// Check if destination already has backups (potential conflicts)
	destPath := fmt.Sprintf("%s/metadata/%s/backups", m.mountPath, m.toPath)
	destSecret, err := m.client.Logical().ListWithContext(ctx, destPath)
	if err == nil && destSecret != nil {
		if keys, ok := destSecret.Data["keys"].([]interface{}); ok && len(keys) > 0 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Destination already contains %d backups - potential naming conflicts", len(keys)))

			// Check for specific conflicts
			destBackups := make(map[string]bool)
			for _, key := range keys {
				if keyStr, ok := key.(string); ok {
					destBackups[keyStr] = true
				}
			}

			conflicts := []string{}
			for _, sourceBackup := range sourceBackups {
				if destBackups[sourceBackup] {
					conflicts = append(conflicts, sourceBackup)
				}
			}

			if len(conflicts) > 0 {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Backup name conflicts detected: %s", strings.Join(conflicts, ", ")))
			}
		}
	}

	// Add strategy-specific warnings
	migrationInfo := GetMigrationInfo(m.fromStrategy, m.toStrategy, m.fromPath, m.toPath)
	result.Warnings = append(result.Warnings, migrationInfo.Risks...)
	result.Benefits = migrationInfo.Benefits

	result.SourceBackupCount = len(sourceBackups)
	return result, nil
}

// MigrationResult contains the results of a migration operation
type MigrationResult struct {
	FromStrategy    StorageStrategy
	ToStrategy      StorageStrategy
	FromPath        string
	ToPath          string
	TotalBackups    int
	MigratedBackups []string
	FailedBackups   []string
	DryRun          bool
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
}

// ValidationResult contains the results of migration validation
type ValidationResult struct {
	Valid             bool
	Warnings          []string
	Errors            []string
	Benefits          []string
	SourceBackupCount int
}

// GetMigrationSummary returns a human-readable summary of the migration result
func (r *MigrationResult) GetMigrationSummary() string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("Migration from %s to %s:\n", r.FromStrategy, r.ToStrategy))
	summary.WriteString(fmt.Sprintf("  Source path: %s\n", r.FromPath))
	summary.WriteString(fmt.Sprintf("  Destination path: %s\n", r.ToPath))
	summary.WriteString(fmt.Sprintf("  Total backups: %d\n", r.TotalBackups))
	summary.WriteString(fmt.Sprintf("  Successfully migrated: %d\n", len(r.MigratedBackups)))
	summary.WriteString(fmt.Sprintf("  Failed: %d\n", len(r.FailedBackups)))
	summary.WriteString(fmt.Sprintf("  Duration: %v\n", r.Duration))

	if r.DryRun {
		summary.WriteString("  [DRY RUN] - No actual changes made\n")
	}

	if len(r.FailedBackups) > 0 {
		summary.WriteString(fmt.Sprintf("  Failed backups: %s\n", strings.Join(r.FailedBackups, ", ")))
	}

	return summary.String()
}
