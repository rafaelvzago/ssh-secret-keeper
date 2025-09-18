package update

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		// Basic version comparisons
		{"Equal versions", "1.0.0", "1.0.0", 0},
		{"v1 newer patch", "1.0.1", "1.0.0", 1},
		{"v2 newer patch", "1.0.0", "1.0.1", -1},
		{"v1 newer minor", "1.1.0", "1.0.0", 1},
		{"v2 newer minor", "1.0.0", "1.1.0", -1},
		{"v1 newer major", "2.0.0", "1.0.0", 1},
		{"v2 newer major", "1.0.0", "2.0.0", -1},

		// With v prefix
		{"With v prefix equal", "v1.0.0", "v1.0.0", 0},
		{"With v prefix v1 newer", "v1.0.1", "v1.0.0", 1},
		{"Mixed v prefix", "v1.0.1", "1.0.0", 1},

		// Pre-release versions
		{"Pre-release vs stable", "1.0.0-beta", "1.0.0", -1},
		{"Stable vs pre-release", "1.0.0", "1.0.0-beta", 1},
		{"Pre-release comparison", "1.0.0-beta.2", "1.0.0-beta.1", 1},
		{"Pre-release alpha vs beta", "1.0.0-beta", "1.0.0-alpha", 1},

		// Different lengths
		{"Missing patch v1", "1.0", "1.0.0", 0},
		{"Missing patch v2", "1.0.0", "1.0", 0},
		{"Missing minor and patch", "2", "1.9.9", 1},

		// Edge cases
		{"Dev version", "dev", "1.0.0", -1},
		{"Empty vs version", "", "1.0.0", -1},
		{"Large numbers", "10.20.30", "9.99.99", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareVersions(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("CompareVersions(%s, %s) = %d, want %d",
					tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected bool
	}{
		{"Newer patch", "1.0.1", "1.0.0", true},
		{"Older patch", "1.0.0", "1.0.1", false},
		{"Equal", "1.0.0", "1.0.0", false},
		{"Newer major", "2.0.0", "1.9.9", true},
		{"Pre-release", "1.0.0", "1.0.0-beta", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNewer(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("IsNewer(%s, %s) = %v, want %v",
					tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Add v prefix", "1.0.0", "v1.0.0"},
		{"Keep v prefix", "v1.0.0", "v1.0.0"},
		{"Dev version", "dev", "dev"},
		{"Empty version", "", ""},
		{"Capital V", "V1.0.0", "v1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatVersion(tt.input)
			if result != tt.expected {
				t.Errorf("FormatVersion(%s) = %s, want %s",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"Valid semver", "1.0.0", false},
		{"Valid with v", "v1.0.0", false},
		{"Valid major.minor", "1.0", false},
		{"Dev version", "dev", false},
		{"Latest version", "latest", false},
		{"Invalid empty", "", true},
		{"Invalid single number", "1", true},
		{"Invalid format", "not-a-version", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVersion(%s) error = %v, wantErr %v",
					tt.version, err, tt.wantErr)
			}
		})
	}
}
