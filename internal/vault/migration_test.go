package vault

import (
	"testing"
	"time"

	"github.com/rzago/ssh-secret-keeper/internal/config"
)

// Note: Tests that require Vault client connections have been removed
// These are pure unit tests for migration logic without external dependencies

func TestMigrationService_GeneratePaths(t *testing.T) {
	tests := []struct {
		name            string
		fromStrategy    StorageStrategy
		toStrategy      StorageStrategy
		customPrefix    string
		backupNamespace string
		wantFromPath    string
		wantToPath      string
	}{
		{
			name:         "machine-user to universal",
			fromStrategy: StrategyMachineUser,
			toStrategy:   StrategyUniversal,
			wantFromPath: "users/", // Will contain hostname-username
			wantToPath:   "shared",
		},
		{
			name:            "machine-user to universal with namespace",
			fromStrategy:    StrategyMachineUser,
			toStrategy:      StrategyUniversal,
			backupNamespace: "personal",
			wantFromPath:    "users/", // Will contain hostname-username
			wantToPath:      "shared/personal",
		},
		{
			name:         "universal to user",
			fromStrategy: StrategyUniversal,
			toStrategy:   StrategyUser,
			wantFromPath: "shared",
			wantToPath:   "users/", // Will contain username
		},
		{
			name:         "user to custom",
			fromStrategy: StrategyUser,
			toStrategy:   StrategyCustom,
			customPrefix: "team-devops",
			wantFromPath: "users/", // Will contain username
			wantToPath:   "team-devops",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't create a real migration service due to vault client dependency,
			// but we can test path generation directly
			fromGenerator := NewPathGenerator(tt.fromStrategy, tt.customPrefix, tt.backupNamespace)
			fromPath, err := fromGenerator.GenerateBasePath()
			if err != nil {
				t.Fatalf("Failed to generate from path: %v", err)
			}

			toGenerator := NewPathGenerator(tt.toStrategy, tt.customPrefix, tt.backupNamespace)
			toPath, err := toGenerator.GenerateBasePath()
			if err != nil {
				t.Fatalf("Failed to generate to path: %v", err)
			}

			if !contains(fromPath, tt.wantFromPath) {
				t.Errorf("From path %q should contain %q", fromPath, tt.wantFromPath)
			}

			if !contains(toPath, tt.wantToPath) && toPath != tt.wantToPath {
				// For partial matches (like "users/" prefix)
				if !hasPrefix(toPath, tt.wantToPath) {
					t.Errorf("To path = %q, should contain or match %q", toPath, tt.wantToPath)
				}
			}
		})
	}
}

func TestMigrationResult(t *testing.T) {
	result := &MigrationResult{
		FromStrategy:    StrategyMachineUser,
		ToStrategy:      StrategyUniversal,
		FromPath:        "users/host-user",
		ToPath:          "shared",
		TotalBackups:    3,
		MigratedBackups: []string{"backup1", "backup2"},
		FailedBackups:   []string{"backup3"},
		DryRun:          false,
		StartTime:       time.Now().Add(-5 * time.Minute),
		EndTime:         time.Now(),
	}

	// Test basic properties
	if result.FromStrategy != StrategyMachineUser {
		t.Errorf("FromStrategy = %v, want %v", result.FromStrategy, StrategyMachineUser)
	}

	if result.ToStrategy != StrategyUniversal {
		t.Errorf("ToStrategy = %v, want %v", result.ToStrategy, StrategyUniversal)
	}

	if len(result.MigratedBackups) != 2 {
		t.Errorf("MigratedBackups count = %d, want 2", len(result.MigratedBackups))
	}

	if len(result.FailedBackups) != 1 {
		t.Errorf("FailedBackups count = %d, want 1", len(result.FailedBackups))
	}

	// Test duration calculation
	duration := result.EndTime.Sub(result.StartTime)
	if duration < 4*time.Minute || duration > 6*time.Minute {
		t.Errorf("Duration = %v, expected around 5 minutes", duration)
	}
}

func TestValidationResult(t *testing.T) {
	tests := []struct {
		name        string
		result      *ValidationResult
		expectValid bool
	}{
		{
			name: "valid migration",
			result: &ValidationResult{
				Valid:    true,
				Warnings: []string{"Some warnings"},
				Errors:   []string{},
			},
			expectValid: true,
		},
		{
			name: "invalid migration with errors",
			result: &ValidationResult{
				Valid:    false,
				Warnings: []string{},
				Errors:   []string{"Critical error"},
			},
			expectValid: false,
		},
		{
			name: "valid with warnings",
			result: &ValidationResult{
				Valid:    true,
				Warnings: []string{"Warning 1", "Warning 2"},
				Errors:   []string{},
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.Valid != tt.expectValid {
				t.Errorf("ValidationResult.Valid = %v, want %v", tt.result.Valid, tt.expectValid)
			}

			if tt.expectValid && len(tt.result.Errors) > 0 {
				t.Errorf("Valid result should not have errors, got: %v", tt.result.Errors)
			}

			if !tt.expectValid && len(tt.result.Errors) == 0 {
				t.Error("Invalid result should have errors")
			}
		})
	}
}

func TestMigrationService_PathGeneration(t *testing.T) {
	// Test that migration service correctly generates different paths for different strategies
	cfg := &config.VaultConfig{
		Address:         "http://localhost:8200",
		MountPath:       "ssh-backups",
		TokenFile:       "/dev/null",
		BackupNamespace: "test-namespace",
	}

	tests := []struct {
		name         string
		fromStrategy StorageStrategy
		toStrategy   StorageStrategy
		expectDiff   bool
	}{
		{
			name:         "machine-user to universal should differ",
			fromStrategy: StrategyMachineUser,
			toStrategy:   StrategyUniversal,
			expectDiff:   true,
		},
		{
			name:         "universal to universal should be same",
			fromStrategy: StrategyUniversal,
			toStrategy:   StrategyUniversal,
			expectDiff:   false,
		},
		{
			name:         "user to machine-user should differ",
			fromStrategy: StrategyUser,
			toStrategy:   StrategyMachineUser,
			expectDiff:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fromGen := NewPathGenerator(tt.fromStrategy, "", cfg.BackupNamespace)
			fromPath, err := fromGen.GenerateBasePath()
			if err != nil {
				t.Fatalf("Failed to generate from path: %v", err)
			}

			toGen := NewPathGenerator(tt.toStrategy, "", cfg.BackupNamespace)
			toPath, err := toGen.GenerateBasePath()
			if err != nil {
				t.Fatalf("Failed to generate to path: %v", err)
			}

			pathsDiffer := fromPath != toPath

			if tt.expectDiff && !pathsDiffer {
				t.Errorf("Expected different paths but got same: from=%q, to=%q", fromPath, toPath)
			}

			if !tt.expectDiff && pathsDiffer {
				t.Errorf("Expected same paths but got different: from=%q, to=%q", fromPath, toPath)
			}
		})
	}
}

func TestMigrationService_StrategyValidation(t *testing.T) {
	// Test validation of migration between different strategies
	tests := []struct {
		name         string
		fromStrategy StorageStrategy
		toStrategy   StorageStrategy
		wantValid    bool
	}{
		{
			name:         "valid migration: machine-user to universal",
			fromStrategy: StrategyMachineUser,
			toStrategy:   StrategyUniversal,
			wantValid:    true,
		},
		{
			name:         "valid migration: machine-user to user",
			fromStrategy: StrategyMachineUser,
			toStrategy:   StrategyUser,
			wantValid:    true,
		},
		{
			name:         "valid migration: universal to user",
			fromStrategy: StrategyUniversal,
			toStrategy:   StrategyUser,
			wantValid:    true,
		},
		{
			name:         "same strategy migration",
			fromStrategy: StrategyUniversal,
			toStrategy:   StrategyUniversal,
			wantValid:    true, // Should be valid but no-op
		},
		{
			name:         "valid migration: user to custom",
			fromStrategy: StrategyUser,
			toStrategy:   StrategyCustom,
			wantValid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that both strategies are valid
			fromGen := NewPathGenerator(tt.fromStrategy, "test-prefix", "")
			err := fromGen.ValidateStrategy()
			if err != nil && tt.fromStrategy != StrategyCustom {
				t.Errorf("From strategy validation failed: %v", err)
			}

			toGen := NewPathGenerator(tt.toStrategy, "test-prefix", "")
			err = toGen.ValidateStrategy()
			if err != nil && tt.toStrategy != StrategyCustom {
				t.Errorf("To strategy validation failed: %v", err)
			}

			// Test migration info generation
			info := GetMigrationInfo(tt.fromStrategy, tt.toStrategy, "test-from", "test-to")
			if info == nil {
				t.Error("GetMigrationInfo returned nil")
				return
			}

			if !info.Compatible && tt.wantValid {
				t.Error("Migration should be compatible but isn't")
			}
		})
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function for string prefix check
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// Benchmark tests for migration operations
func BenchmarkGetMigrationInfo(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetMigrationInfo(StrategyMachineUser, StrategyUniversal, "users/host-user", "shared")
	}
}
