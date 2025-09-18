package vault

import (
	"fmt"
	"strings"
	"testing"
)

// Note: All tests that require Vault client connections have been removed
// These are pure unit tests for error handling logic without external dependencies

func TestPathGenerator_ErrorScenarios(t *testing.T) {
	// Test error scenarios for path generation across all strategies

	tests := []struct {
		name         string
		strategy     StorageStrategy
		customPrefix string
		namespace    string
		wantErr      bool
		errContains  string
		description  string
	}{
		{
			name:         "custom strategy without prefix",
			strategy:     StrategyCustom,
			customPrefix: "",
			wantErr:      true,
			errContains:  "custom prefix is required",
			description:  "Custom strategy must have a prefix",
		},
		{
			name:         "custom strategy with empty prefix",
			strategy:     StrategyCustom,
			customPrefix: "   ", // Whitespace only
			wantErr:      false, // Whitespace is actually valid, gets sanitized
			description:  "Custom strategy prefix with whitespace gets sanitized",
		},
		{
			name:         "custom strategy with path separators",
			strategy:     StrategyCustom,
			customPrefix: "team/devops",
			wantErr:      true,
			errContains:  "path separators",
			description:  "Custom strategy prefix cannot contain slashes",
		},
		{
			name:         "custom strategy with invalid characters",
			strategy:     StrategyCustom,
			customPrefix: "team\\devops",
			wantErr:      false, // Backslashes get sanitized, not rejected
			description:  "Custom strategy prefix with backslashes gets sanitized",
		},
		{
			name:        "unknown strategy",
			strategy:    StorageStrategy("unknown"),
			wantErr:     true,
			errContains: "unknown storage strategy",
			description: "Unknown strategies should be rejected",
		},
		{
			name:        "valid universal strategy",
			strategy:    StrategyUniversal,
			namespace:   "test-namespace",
			wantErr:     false,
			description: "Universal strategy should work with namespace",
		},
		{
			name:        "valid user strategy",
			strategy:    StrategyUser,
			wantErr:     false,
			description: "User strategy should work without additional config",
		},
		{
			name:        "valid machine-user strategy",
			strategy:    StrategyMachineUser,
			wantErr:     false,
			description: "Machine-user strategy should work without additional config",
		},
		{
			name:         "valid custom strategy",
			strategy:     StrategyCustom,
			customPrefix: "team-devops",
			wantErr:      false,
			description:  "Custom strategy should work with valid prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pathGenerator := NewPathGenerator(tt.strategy, tt.customPrefix, tt.namespace)

			// Test validation
			err := pathGenerator.ValidateStrategy()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateStrategy() expected error for %s", tt.description)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got: %v", tt.errContains, err)
				}
				t.Logf("Expected validation error: %v (%s)", err, tt.description)
				return
			}

			if err != nil {
				t.Errorf("ValidateStrategy() unexpected error for %s: %v", tt.description, err)
				return
			}

			// Test path generation
			path, err := pathGenerator.GenerateBasePath()
			if err != nil {
				t.Errorf("GenerateBasePath() unexpected error for %s: %v", tt.description, err)
				return
			}

			if path == "" {
				t.Errorf("GenerateBasePath() returned empty path for %s", tt.description)
			}

			t.Logf("Generated path: %s (%s)", path, tt.description)
		})
	}
}

func TestStorageStrategy_ParseErrors(t *testing.T) {
	// Test error scenarios for strategy parsing

	tests := []struct {
		input       string
		wantErr     bool
		errContains string
		description string
	}{
		{
			input:       "",
			wantErr:     true,
			errContains: "invalid storage strategy",
			description: "Empty strategy string should be rejected",
		},
		{
			input:       "   ",
			wantErr:     true,
			errContains: "invalid storage strategy",
			description: "Whitespace-only strategy should be rejected",
		},
		{
			input:       "invalid-strategy",
			wantErr:     true,
			errContains: "invalid storage strategy",
			description: "Unknown strategy should be rejected",
		},
		{
			input:       "INVALID",
			wantErr:     true,
			errContains: "invalid storage strategy",
			description: "Unknown strategy in caps should be rejected",
		},
		{
			input:       "universal123",
			wantErr:     true,
			errContains: "invalid storage strategy",
			description: "Strategy with numbers should be rejected",
		},
		{
			input:       "universal-extra",
			wantErr:     true,
			errContains: "invalid storage strategy",
			description: "Strategy with extra suffix should be rejected",
		},
		{
			input:       "universal",
			wantErr:     false,
			description: "Valid universal strategy should be accepted",
		},
		{
			input:       "UNIVERSAL",
			wantErr:     false,
			description: "Case insensitive universal strategy should be accepted",
		},
		{
			input:       " user ",
			wantErr:     false,
			description: "Strategy with whitespace should be trimmed and accepted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			strategy, err := ParseStrategy(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseStrategy(%q) expected error but got none", tt.input)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got: %v", tt.errContains, err)
				}
				t.Logf("Expected error: %v", err)
				return
			}

			if err != nil {
				t.Errorf("ParseStrategy(%q) unexpected error: %v", tt.input, err)
				return
			}

			if strategy == "" {
				t.Errorf("ParseStrategy(%q) returned empty strategy", tt.input)
			}

			t.Logf("Parsed strategy: %s", strategy)
		})
	}
}

func TestConcurrentStrategyOperations(t *testing.T) {
	// Test that strategy operations are safe under concurrent access
	// This is important for CLI tools that might be run concurrently

	strategies := []StorageStrategy{
		StrategyUniversal,
		StrategyUser,
		StrategyMachineUser,
		StrategyCustom,
	}

	// Test concurrent path generation
	t.Run("concurrent_path_generation", func(t *testing.T) {
		done := make(chan bool)
		errors := make(chan error, len(strategies)*10)

		for _, strategy := range strategies {
			for i := 0; i < 10; i++ {
				go func(s StorageStrategy, iteration int) {
					defer func() { done <- true }()

					customPrefix := ""
					if s == StrategyCustom {
						customPrefix = "test-prefix"
					}

					pathGen := NewPathGenerator(s, customPrefix, "test-namespace")

					// Validate strategy
					if err := pathGen.ValidateStrategy(); err != nil {
						if s != StrategyCustom || customPrefix != "" {
							errors <- err
							return
						}
					}

					// Generate path
					path, err := pathGen.GenerateBasePath()
					if err != nil {
						errors <- err
						return
					}

					if path == "" {
						errors <- fmt.Errorf("empty path generated")
						return
					}
				}(strategy, i)
			}
		}

		// Wait for all goroutines
		for i := 0; i < len(strategies)*10; i++ {
			<-done
		}

		// Check for errors
		close(errors)
		for err := range errors {
			t.Errorf("Concurrent operation error: %v", err)
		}
	})

	// Test concurrent strategy parsing
	t.Run("concurrent_strategy_parsing", func(t *testing.T) {
		strategyStrings := []string{"universal", "user", "machine-user", "custom"}
		done := make(chan bool)
		errors := make(chan error, len(strategyStrings)*10)

		for _, strategyStr := range strategyStrings {
			for i := 0; i < 10; i++ {
				go func(s string) {
					defer func() { done <- true }()

					strategy, err := ParseStrategy(s)
					if err != nil {
						errors <- err
						return
					}

					if strategy == "" {
						errors <- fmt.Errorf("empty strategy parsed")
						return
					}
				}(strategyStr)
			}
		}

		// Wait for all goroutines
		for i := 0; i < len(strategyStrings)*10; i++ {
			<-done
		}

		// Check for errors
		close(errors)
		for err := range errors {
			t.Errorf("Concurrent parsing error: %v", err)
		}
	})
}
