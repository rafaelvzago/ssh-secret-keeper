package vault

import (
	"os"
	"strings"
	"testing"
)

func TestNewPathGenerator(t *testing.T) {
	tests := []struct {
		name         string
		strategy     StorageStrategy
		customPrefix string
		namespace    string
	}{
		{
			name:     "universal strategy",
			strategy: StrategyUniversal,
		},
		{
			name:      "universal with namespace",
			strategy:  StrategyUniversal,
			namespace: "personal",
		},
		{
			name:     "user strategy",
			strategy: StrategyUser,
		},
		{
			name:     "machine-user strategy",
			strategy: StrategyMachineUser,
		},
		{
			name:         "custom strategy",
			strategy:     StrategyCustom,
			customPrefix: "team-devops",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewPathGenerator(tt.strategy, tt.customPrefix, tt.namespace)

			if generator == nil {
				t.Error("NewPathGenerator() returned nil")
				return
			}

			if generator.strategy != tt.strategy {
				t.Errorf("NewPathGenerator() strategy = %v, want %v", generator.strategy, tt.strategy)
			}

			if generator.customPrefix != tt.customPrefix {
				t.Errorf("NewPathGenerator() customPrefix = %v, want %v", generator.customPrefix, tt.customPrefix)
			}

			if generator.namespace != tt.namespace {
				t.Errorf("NewPathGenerator() namespace = %v, want %v", generator.namespace, tt.namespace)
			}
		})
	}
}

func TestPathGenerator_GenerateBasePath(t *testing.T) {
	// Set up test environment variables
	originalUser := os.Getenv("USER")
	originalUsername := os.Getenv("USERNAME")
	defer func() {
		os.Setenv("USER", originalUser)
		os.Setenv("USERNAME", originalUsername)
	}()

	os.Setenv("USER", "testuser")
	os.Setenv("USERNAME", "testuser")

	tests := []struct {
		name         string
		strategy     StorageStrategy
		customPrefix string
		namespace    string
		wantContains string
		wantErr      bool
	}{
		{
			name:         "universal without namespace",
			strategy:     StrategyUniversal,
			wantContains: "shared",
			wantErr:      false,
		},
		{
			name:         "universal with namespace",
			strategy:     StrategyUniversal,
			namespace:    "personal",
			wantContains: "shared/personal",
			wantErr:      false,
		},
		{
			name:         "user strategy",
			strategy:     StrategyUser,
			wantContains: "users/testuser",
			wantErr:      false,
		},
		{
			name:         "machine-user strategy",
			strategy:     StrategyMachineUser,
			wantContains: "users/",
			wantErr:      false,
		},
		{
			name:         "custom strategy valid",
			strategy:     StrategyCustom,
			customPrefix: "team-devops",
			wantContains: "team-devops",
			wantErr:      false,
		},
		{
			name:     "custom strategy missing prefix",
			strategy: StrategyCustom,
			wantErr:  true,
		},
		{
			name:     "invalid strategy",
			strategy: StorageStrategy("invalid"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewPathGenerator(tt.strategy, tt.customPrefix, tt.namespace)

			result, err := generator.GenerateBasePath()

			if tt.wantErr {
				if err == nil {
					t.Error("GenerateBasePath() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GenerateBasePath() unexpected error: %v", err)
				return
			}

			if result == "" {
				t.Error("GenerateBasePath() returned empty string")
				return
			}

			if !strings.Contains(result, tt.wantContains) {
				t.Errorf("GenerateBasePath() = %q, should contain %q", result, tt.wantContains)
			}
		})
	}
}

func TestPathGenerator_ValidateStrategy(t *testing.T) {
	tests := []struct {
		name         string
		strategy     StorageStrategy
		customPrefix string
		wantErr      bool
		errContains  string
	}{
		{
			name:     "universal strategy valid",
			strategy: StrategyUniversal,
			wantErr:  false,
		},
		{
			name:     "user strategy valid",
			strategy: StrategyUser,
			wantErr:  false,
		},
		{
			name:     "machine-user strategy valid",
			strategy: StrategyMachineUser,
			wantErr:  false,
		},
		{
			name:         "custom strategy valid",
			strategy:     StrategyCustom,
			customPrefix: "team-devops",
			wantErr:      false,
		},
		{
			name:        "custom strategy missing prefix",
			strategy:    StrategyCustom,
			wantErr:     true,
			errContains: "custom prefix is required",
		},
		{
			name:         "custom strategy with slash",
			strategy:     StrategyCustom,
			customPrefix: "team/devops",
			wantErr:      true,
			errContains:  "cannot contain path separators",
		},
		{
			name:        "unknown strategy",
			strategy:    StorageStrategy("unknown"),
			wantErr:     true,
			errContains: "unknown storage strategy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewPathGenerator(tt.strategy, tt.customPrefix, "")

			err := generator.ValidateStrategy()

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateStrategy() expected error but got none")
					return
				}

				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateStrategy() error = %q, should contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateStrategy() unexpected error: %v", err)
			}
		})
	}
}

func TestParseStrategy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected StorageStrategy
		wantErr  bool
	}{
		{
			name:     "universal",
			input:    "universal",
			expected: StrategyUniversal,
			wantErr:  false,
		},
		{
			name:     "shared alias",
			input:    "shared",
			expected: StrategyUniversal,
			wantErr:  false,
		},
		{
			name:     "user",
			input:    "user",
			expected: StrategyUser,
			wantErr:  false,
		},
		{
			name:     "machine-user",
			input:    "machine-user",
			expected: StrategyMachineUser,
			wantErr:  false,
		},
		{
			name:     "machine_user alias",
			input:    "machine_user",
			expected: StrategyMachineUser,
			wantErr:  false,
		},
		{
			name:     "legacy alias",
			input:    "legacy",
			expected: StrategyMachineUser,
			wantErr:  false,
		},
		{
			name:     "custom",
			input:    "custom",
			expected: StrategyCustom,
			wantErr:  false,
		},
		{
			name:     "case insensitive",
			input:    "UNIVERSAL",
			expected: StrategyUniversal,
			wantErr:  false,
		},
		{
			name:     "with whitespace",
			input:    " user ",
			expected: StrategyUser,
			wantErr:  false,
		},
		{
			name:    "invalid strategy",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseStrategy(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseStrategy() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseStrategy() unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("ParseStrategy() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetAllStrategies(t *testing.T) {
	strategies := GetAllStrategies()

	expectedStrategies := []StorageStrategy{
		StrategyUniversal,
		StrategyUser,
		StrategyMachineUser,
		StrategyCustom,
	}

	if len(strategies) != len(expectedStrategies) {
		t.Errorf("GetAllStrategies() returned %d strategies, want %d", len(strategies), len(expectedStrategies))
	}

	for _, expected := range expectedStrategies {
		if description, exists := strategies[expected]; !exists {
			t.Errorf("GetAllStrategies() missing strategy %v", expected)
		} else if description == "" {
			t.Errorf("GetAllStrategies() strategy %v has empty description", expected)
		}
	}
}

func TestPathGenerator_GetStrategyDescription(t *testing.T) {
	tests := []struct {
		name         string
		strategy     StorageStrategy
		customPrefix string
		namespace    string
		wantContains string
	}{
		{
			name:         "universal without namespace",
			strategy:     StrategyUniversal,
			wantContains: "Universal storage",
		},
		{
			name:         "universal with namespace",
			strategy:     StrategyUniversal,
			namespace:    "personal",
			wantContains: "namespace 'personal'",
		},
		{
			name:         "user strategy",
			strategy:     StrategyUser,
			wantContains: "User-scoped storage",
		},
		{
			name:         "machine-user strategy",
			strategy:     StrategyMachineUser,
			wantContains: "Machine-user scoped storage",
		},
		{
			name:         "custom strategy",
			strategy:     StrategyCustom,
			customPrefix: "team-devops",
			wantContains: "prefix 'team-devops'",
		},
		{
			name:         "unknown strategy",
			strategy:     StorageStrategy("unknown"),
			wantContains: "Unknown storage strategy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewPathGenerator(tt.strategy, tt.customPrefix, tt.namespace)

			description := generator.GetStrategyDescription()

			if description == "" {
				t.Error("GetStrategyDescription() returned empty string")
				return
			}

			if !strings.Contains(description, tt.wantContains) {
				t.Errorf("GetStrategyDescription() = %q, should contain %q", description, tt.wantContains)
			}
		})
	}
}

func TestGetMigrationInfo(t *testing.T) {
	tests := []struct {
		name         string
		fromStrategy StorageStrategy
		toStrategy   StorageStrategy
		fromPath     string
		toPath       string
		wantBenefits bool
		wantRisks    bool
	}{
		{
			name:         "machine-user to universal",
			fromStrategy: StrategyMachineUser,
			toStrategy:   StrategyUniversal,
			fromPath:     "users/host-user",
			toPath:       "shared",
			wantBenefits: true,
			wantRisks:    true,
		},
		{
			name:         "machine-user to user",
			fromStrategy: StrategyMachineUser,
			toStrategy:   StrategyUser,
			fromPath:     "users/host-user",
			toPath:       "users/user",
			wantBenefits: true,
			wantRisks:    true,
		},
		{
			name:         "universal to user",
			fromStrategy: StrategyUniversal,
			toStrategy:   StrategyUser,
			fromPath:     "shared",
			toPath:       "users/user",
			wantBenefits: true,
			wantRisks:    true,
		},
		{
			name:         "same strategy",
			fromStrategy: StrategyUniversal,
			toStrategy:   StrategyUniversal,
			fromPath:     "shared",
			toPath:       "shared",
			wantBenefits: true,
			wantRisks:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := GetMigrationInfo(tt.fromStrategy, tt.toStrategy, tt.fromPath, tt.toPath)

			if info == nil {
				t.Error("GetMigrationInfo() returned nil")
				return
			}

			if info.FromStrategy != tt.fromStrategy {
				t.Errorf("GetMigrationInfo() FromStrategy = %v, want %v", info.FromStrategy, tt.fromStrategy)
			}

			if info.ToStrategy != tt.toStrategy {
				t.Errorf("GetMigrationInfo() ToStrategy = %v, want %v", info.ToStrategy, tt.toStrategy)
			}

			if info.FromPath != tt.fromPath {
				t.Errorf("GetMigrationInfo() FromPath = %q, want %q", info.FromPath, tt.fromPath)
			}

			if info.ToPath != tt.toPath {
				t.Errorf("GetMigrationInfo() ToPath = %q, want %q", info.ToPath, tt.toPath)
			}

			if tt.wantBenefits && len(info.Benefits) == 0 {
				t.Error("GetMigrationInfo() expected benefits but got none")
			}

			if tt.wantRisks && len(info.Risks) == 0 {
				t.Error("GetMigrationInfo() expected risks but got none")
			}
		})
	}
}

func TestPathGenerator_CrossMachineCompatibility(t *testing.T) {
	// Test that different machines generate different machine-user paths
	// but same paths for other strategies

	// Set up test environment
	originalUser := os.Getenv("USER")
	defer os.Setenv("USER", originalUser)
	os.Setenv("USER", "testuser")

	tests := []struct {
		name            string
		strategy        StorageStrategy
		expectDifferent bool // Whether different hostnames should produce different paths
	}{
		{
			name:            "universal strategy",
			strategy:        StrategyUniversal,
			expectDifferent: false, // Same path across machines
		},
		{
			name:            "user strategy",
			strategy:        StrategyUser,
			expectDifferent: false, // Same path across machines for same user
		},
		{
			name:            "machine-user strategy",
			strategy:        StrategyMachineUser,
			expectDifferent: true, // Different path per machine
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator1 := NewPathGenerator(tt.strategy, "", "")
			generator2 := NewPathGenerator(tt.strategy, "", "")

			path1, err1 := generator1.GenerateBasePath()
			path2, err2 := generator2.GenerateBasePath()

			if err1 != nil || err2 != nil {
				t.Fatalf("GenerateBasePath() errors: %v, %v", err1, err2)
			}

			// For this test, both generators will produce the same result
			// since they're running on the same machine. The real test
			// would be across actual different machines.
			if path1 != path2 {
				t.Errorf("Expected same paths on same machine, got %q and %q", path1, path2)
			}

			// Verify path format expectations
			switch tt.strategy {
			case StrategyUniversal:
				if !strings.Contains(path1, "shared") {
					t.Errorf("Universal strategy path should contain 'shared': %q", path1)
				}
			case StrategyUser:
				if !strings.Contains(path1, "users/testuser") {
					t.Errorf("User strategy path should contain 'users/testuser': %q", path1)
				}
			case StrategyMachineUser:
				if !strings.Contains(path1, "users/") || !strings.Contains(path1, "testuser") {
					t.Errorf("Machine-user strategy path should contain hostname and username: %q", path1)
				}
			}
		})
	}
}

func BenchmarkPathGenerator_GenerateBasePath(b *testing.B) {
	generator := NewPathGenerator(StrategyMachineUser, "", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateBasePath()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseStrategy(b *testing.B) {
	testStrategy := "universal"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseStrategy(testStrategy)
		if err != nil {
			b.Fatal(err)
		}
	}
}
