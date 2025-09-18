package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPathNormalizer_NormalizePath(t *testing.T) {
	normalizer := NewPathNormalizer()

	tests := []struct {
		name     string
		input    string
		expected string
		skipOS   string // Skip test on specific OS
	}{
		{
			name:     "already relative path",
			input:    "~/.ssh",
			expected: "~/.ssh",
		},
		{
			name:     "home directory only",
			input:    "~/",
			expected: "~",
		},
		{
			name:     "linux home path",
			input:    "/home/user/.ssh",
			expected: "~/.ssh",
			skipOS:   "windows",
		},
		{
			name:     "linux home root",
			input:    "/home/user",
			expected: "~",
			skipOS:   "windows",
		},
		{
			name:     "linux root user",
			input:    "/root/.ssh",
			expected: "~/.ssh",
			skipOS:   "windows",
		},
		{
			name:     "linux root home",
			input:    "/root",
			expected: "~",
			skipOS:   "windows",
		},
		{
			name:     "absolute path outside home",
			input:    "/etc/ssh",
			expected: "/etc/ssh",
			skipOS:   "windows",
		},
		{
			name:     "windows home path",
			input:    "C:\\Users\\user\\.ssh",
			expected: "~\\.ssh",
			skipOS:   "linux",
		},
		{
			name:     "windows home root",
			input:    "C:\\Users\\user",
			expected: "~",
			skipOS:   "linux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOS == runtime.GOOS {
				t.Skipf("Skipping test on %s", runtime.GOOS)
			}

			result, err := normalizer.NormalizePath(tt.input)
			if err != nil {
				t.Errorf("NormalizePath() unexpected error: %v", err)
				return
			}

			// Clean paths for comparison
			expected := filepath.Clean(tt.expected)
			result = filepath.Clean(result)

			if result != expected {
				t.Errorf("NormalizePath() = %q, want %q", result, expected)
			}
		})
	}
}

func TestPathNormalizer_ResolvePath(t *testing.T) {
	normalizer := NewPathNormalizer()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "relative path with tilde",
			input:   "~/.ssh",
			wantErr: false,
		},
		{
			name:    "home directory only",
			input:   "~",
			wantErr: false,
		},
		{
			name:    "absolute path",
			input:   "/tmp/test",
			wantErr: false,
		},
		{
			name:    "relative path",
			input:   "test/path",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizer.ResolvePath(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("ResolvePath() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ResolvePath() unexpected error: %v", err)
				return
			}

			if result == "" {
				t.Error("ResolvePath() returned empty string")
			}

			// For tilde paths, result should be absolute
			if tt.input[0] == '~' {
				if !filepath.IsAbs(result) {
					t.Errorf("ResolvePath() = %q, expected absolute path", result)
				}
			}
		})
	}
}

func TestPathNormalizer_DetectHomeDirectoryPattern(t *testing.T) {
	normalizer := NewPathNormalizer()

	tests := []struct {
		name     string
		input    string
		expected string
		testOS   string // Only test on specific OS
	}{
		{
			name:     "linux home pattern",
			input:    "/home/user/.ssh",
			expected: "/home/",
			testOS:   "linux",
		},
		{
			name:     "linux root pattern",
			input:    "/root/.bashrc",
			expected: "/root",
			testOS:   "linux",
		},
		{
			name:     "macos home pattern",
			input:    "/Users/user/Documents",
			expected: "/Users/",
			testOS:   "darwin",
		},
		{
			name:     "windows home pattern",
			input:    "C:\\Users\\user\\Documents",
			expected: "\\Users\\",
			testOS:   "windows",
		},
		{
			name:     "no pattern match",
			input:    "/var/log/test",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.testOS != "" && tt.testOS != runtime.GOOS {
				t.Skipf("Skipping test for %s on %s", tt.testOS, runtime.GOOS)
			}

			result := normalizer.DetectHomeDirectoryPattern(tt.input)
			if result != tt.expected {
				t.Errorf("DetectHomeDirectoryPattern() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPathNormalizer_IsRelativePath(t *testing.T) {
	normalizer := NewPathNormalizer()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "tilde with path",
			input:    "~/.ssh",
			expected: true,
		},
		{
			name:     "tilde only",
			input:    "~",
			expected: true,
		},
		{
			name:     "absolute path",
			input:    "/home/user/.ssh",
			expected: false,
		},
		{
			name:     "relative path without tilde",
			input:    "ssh/keys",
			expected: false,
		},
		{
			name:     "windows absolute path",
			input:    "C:\\Users\\user",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizer.IsRelativePath(tt.input)
			if result != tt.expected {
				t.Errorf("IsRelativePath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSanitizePathComponent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean string",
			input:    "hostname",
			expected: "hostname",
		},
		{
			name:     "with spaces",
			input:    "my hostname",
			expected: "my_hostname",
		},
		{
			name:     "with special characters",
			input:    "host/name:with*chars",
			expected: "host_name_with_chars",
		},
		{
			name:     "with leading/trailing dots and underscores",
			input:    "._hostname._",
			expected: "hostname",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "unknown",
		},
		{
			name:     "only special chars",
			input:    "/*?<>|",
			expected: "unknown",
		},
		{
			name:     "with newlines and tabs",
			input:    "host\nname\ttest",
			expected: "host_name_test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizePathComponent(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizePathComponent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPathNormalizer_CrossUserScenario(t *testing.T) {
	// Test cross-user path normalization scenario
	normalizer := NewPathNormalizer()

	// Simulate different user paths that should normalize to the same relative path
	testPaths := []string{}
	expectedNorm := "~/.ssh"

	switch runtime.GOOS {
	case "linux":
		testPaths = []string{
			"/home/user1/.ssh",
			"/home/user2/.ssh",
			"/home/alice/.ssh",
			"/root/.ssh",
		}
	case "darwin":
		testPaths = []string{
			"/Users/user1/.ssh",
			"/Users/user2/.ssh",
			"/Users/alice/.ssh",
		}
	case "windows":
		testPaths = []string{
			"C:\\Users\\user1\\.ssh",
			"C:\\Users\\user2\\.ssh",
			"C:\\Users\\alice\\.ssh",
		}
		expectedNorm = "~\\.ssh"
	default:
		t.Skip("Unsupported OS for cross-user scenario test")
	}

	for _, path := range testPaths {
		t.Run("normalize_"+path, func(t *testing.T) {
			normalized, err := normalizer.NormalizePath(path)
			if err != nil {
				t.Errorf("NormalizePath() unexpected error: %v", err)
				return
			}

			normalized = filepath.Clean(normalized)
			expected := filepath.Clean(expectedNorm)

			if normalized != expected {
				t.Errorf("NormalizePath(%q) = %q, want %q", path, normalized, expected)
			}
		})
	}
}

func TestPathNormalizer_RoundTrip(t *testing.T) {
	// Test that normalize -> resolve -> normalize gives consistent results
	normalizer := NewPathNormalizer()

	// Get current user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get user home directory")
	}

	sshDir := filepath.Join(homeDir, ".ssh")

	// Normalize the absolute path
	normalized, err := normalizer.NormalizePath(sshDir)
	if err != nil {
		t.Fatalf("NormalizePath() unexpected error: %v", err)
	}

	// Resolve it back to absolute
	resolved, err := normalizer.ResolvePath(normalized)
	if err != nil {
		t.Fatalf("ResolvePath() unexpected error: %v", err)
	}

	// Clean paths for comparison
	originalClean := filepath.Clean(sshDir)
	resolvedClean := filepath.Clean(resolved)

	if originalClean != resolvedClean {
		t.Errorf("Round trip failed: original=%q, resolved=%q", originalClean, resolvedClean)
	}

	// Normalize again - should be the same
	normalizedAgain, err := normalizer.NormalizePath(resolved)
	if err != nil {
		t.Fatalf("Second NormalizePath() unexpected error: %v", err)
	}

	normalizedClean := filepath.Clean(normalized)
	normalizedAgainClean := filepath.Clean(normalizedAgain)

	if normalizedClean != normalizedAgainClean {
		t.Errorf("Double normalization inconsistent: first=%q, second=%q",
			normalizedClean, normalizedAgainClean)
	}
}

func BenchmarkPathNormalizer_NormalizePath(b *testing.B) {
	normalizer := NewPathNormalizer()
	testPath := "/home/user/.ssh"

	if runtime.GOOS == "windows" {
		testPath = "C:\\Users\\user\\.ssh"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := normalizer.NormalizePath(testPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPathNormalizer_ResolvePath(b *testing.B) {
	normalizer := NewPathNormalizer()
	testPath := "~/.ssh"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := normalizer.ResolvePath(testPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSanitizePathComponent(b *testing.B) {
	testComponent := "my-hostname.with-special*chars"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SanitizePathComponent(testComponent)
	}
}
