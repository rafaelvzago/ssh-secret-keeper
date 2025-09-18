package update

import (
	"fmt"
	"strconv"
	"strings"
)

// CompareVersions compares two semantic versions
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func CompareVersions(v1, v2 string) int {
	// Clean version strings
	v1 = cleanVersion(v1)
	v2 = cleanVersion(v2)

	// Split into parts
	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	// Compare major.minor.patch
	for i := 0; i < 3; i++ {
		if i >= len(parts1) && i >= len(parts2) {
			break
		}

		p1 := 0
		if i < len(parts1) {
			p1 = parts1[i]
		}

		p2 := 0
		if i < len(parts2) {
			p2 = parts2[i]
		}

		if p1 > p2 {
			return 1
		} else if p1 < p2 {
			return -1
		}
	}

	// Check for pre-release versions
	pre1 := extractPreRelease(v1)
	pre2 := extractPreRelease(v2)

	// If one has pre-release and other doesn't
	if pre1 != "" && pre2 == "" {
		return -1 // v1 is pre-release, v2 is stable
	}
	if pre1 == "" && pre2 != "" {
		return 1 // v1 is stable, v2 is pre-release
	}

	// Both have pre-release or both don't
	if pre1 != "" && pre2 != "" {
		return strings.Compare(pre1, pre2)
	}

	return 0
}

// IsNewer checks if version v1 is newer than v2
func IsNewer(v1, v2 string) bool {
	return CompareVersions(v1, v2) > 0
}

// cleanVersion removes 'v' prefix and cleans the version string
func cleanVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")
	return version
}

// parseVersion parses a version string into major, minor, patch numbers
func parseVersion(version string) []int {
	// Remove any pre-release or metadata
	if idx := strings.Index(version, "-"); idx != -1 {
		version = version[:idx]
	}
	if idx := strings.Index(version, "+"); idx != -1 {
		version = version[:idx]
	}

	parts := strings.Split(version, ".")
	result := make([]int, 0, 3)

	for i, part := range parts {
		if i >= 3 {
			break // Only consider major.minor.patch
		}

		num, err := strconv.Atoi(part)
		if err != nil {
			num = 0
		}
		result = append(result, num)
	}

	return result
}

// extractPreRelease extracts pre-release identifier from version
func extractPreRelease(version string) string {
	if idx := strings.Index(version, "-"); idx != -1 {
		endIdx := len(version)
		if plusIdx := strings.Index(version[idx:], "+"); plusIdx != -1 {
			endIdx = idx + plusIdx
		}
		return version[idx+1 : endIdx]
	}
	return ""
}

// FormatVersion ensures consistent version formatting
func FormatVersion(version string) string {
	version = cleanVersion(version)
	if !strings.HasPrefix(version, "v") && version != "dev" && version != "" {
		version = "v" + version
	}
	return version
}

// ValidateVersion checks if a version string is valid
func ValidateVersion(version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	if version == "dev" || version == "latest" {
		return nil // Special cases
	}

	cleaned := cleanVersion(version)
	parts := parseVersion(cleaned)

	if len(parts) == 0 {
		return fmt.Errorf("invalid version format: %s", version)
	}

	// Check that we have at least major.minor
	if len(parts) < 2 {
		return fmt.Errorf("version must have at least major.minor: %s", version)
	}

	return nil
}
