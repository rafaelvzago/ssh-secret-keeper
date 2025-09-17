package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PathNormalizer provides cross-platform path normalization functionality
type PathNormalizer struct{}

// NewPathNormalizer creates a new path normalizer instance
func NewPathNormalizer() *PathNormalizer {
	return &PathNormalizer{}
}

// NormalizePath converts absolute paths to relative paths when possible
// This enables cross-user and cross-machine compatibility
func (p *PathNormalizer) NormalizePath(absolutePath string) (string, error) {
	// Handle ~ expansion first - already relative
	if strings.HasPrefix(absolutePath, "~/") {
		return absolutePath, nil
	}

	// Handle Windows home directory patterns
	if runtime.GOOS == "windows" {
		return p.normalizeWindowsPath(absolutePath)
	}

	// Handle Unix-like systems (Linux, macOS)
	return p.normalizeUnixPath(absolutePath)
}

// normalizeUnixPath handles Unix-like path normalization
func (p *PathNormalizer) normalizeUnixPath(absolutePath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return absolutePath, fmt.Errorf("cannot determine home directory: %w", err)
	}

	// Check if path is within current user's home directory
	if strings.HasPrefix(absolutePath, homeDir) {
		relativePath := "~" + strings.TrimPrefix(absolutePath, homeDir)
		return filepath.Clean(relativePath), nil
	}

	// Check for common home directory patterns
	if strings.HasPrefix(absolutePath, "/home/") {
		parts := strings.Split(strings.TrimPrefix(absolutePath, "/home/"), "/")
		if len(parts) >= 1 && parts[0] != "" {
			if len(parts) == 1 {
				// Just /home/username -> ~
				return "~", nil
			}
			// Convert /home/username/... to ~/...
			userPath := strings.Join(parts[1:], "/")
			if userPath == "" {
				return "~", nil
			}
			return filepath.Clean("~/" + userPath), nil
		}
	}

	// Handle root user
	if absolutePath == "/root" {
		return "~", nil
	}
	if strings.HasPrefix(absolutePath, "/root/") {
		userPath := strings.TrimPrefix(absolutePath, "/root/")
		if userPath == "" {
			return "~", nil
		}
		return filepath.Clean("~/" + userPath), nil
	}

	// Path is outside home directory patterns, keep absolute
	return absolutePath, nil
}

// normalizeWindowsPath handles Windows path normalization
func (p *PathNormalizer) normalizeWindowsPath(absolutePath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return absolutePath, fmt.Errorf("cannot determine home directory: %w", err)
	}

	// Normalize path separators
	absolutePath = filepath.Clean(absolutePath)
	homeDir = filepath.Clean(homeDir)

	// Check if path is within current user's home directory
	if strings.HasPrefix(strings.ToLower(absolutePath), strings.ToLower(homeDir)) {
		relativePath := "~" + strings.TrimPrefix(absolutePath, homeDir)
		return filepath.Clean(relativePath), nil
	}

	// Check for common Windows home patterns like C:\Users\username
	if strings.Contains(strings.ToLower(absolutePath), "\\users\\") {
		// Find the Users part and convert to relative path
		parts := strings.Split(absolutePath, "\\")
		for i, part := range parts {
			if strings.ToLower(part) == "users" && i+1 < len(parts) {
				// Skip the username part and create relative path
				if i+2 < len(parts) {
					userPath := strings.Join(parts[i+2:], "\\")
					return filepath.Clean("~/" + userPath), nil
				} else {
					return "~", nil
				}
			}
		}
	}

	// Path is outside home directory patterns, keep absolute
	return absolutePath, nil
}

// ResolvePath converts relative paths to absolute paths for current user
func (p *PathNormalizer) ResolvePath(relativePath string) (string, error) {
	if strings.HasPrefix(relativePath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot resolve home directory: %w", err)
		}
		return filepath.Join(homeDir, relativePath[2:]), nil
	}

	// Handle just ~ (home directory)
	if relativePath == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot resolve home directory: %w", err)
		}
		return homeDir, nil
	}

	// Already absolute or relative to current directory
	return filepath.Abs(relativePath)
}

// DetectHomeDirectoryPattern detects the home directory pattern for cross-platform support
func (p *PathNormalizer) DetectHomeDirectoryPattern(path string) string {
	switch runtime.GOOS {
	case "darwin":
		if strings.HasPrefix(path, "/Users/") {
			return "/Users/"
		}
	case "linux":
		if strings.HasPrefix(path, "/home/") {
			return "/home/"
		}
		if strings.HasPrefix(path, "/root") {
			return "/root"
		}
	case "windows":
		if strings.Contains(strings.ToLower(path), "\\users\\") {
			return "\\Users\\"
		}
	}
	return ""
}

// IsRelativePath checks if a path is relative (starts with ~)
func (p *PathNormalizer) IsRelativePath(path string) bool {
	return strings.HasPrefix(path, "~/") || path == "~"
}

// SanitizePathComponent removes or replaces characters that could cause issues in paths
func SanitizePathComponent(component string) string {
	// Replace problematic characters with underscores
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
		"\t", "_",
		"\n", "_",
		"\r", "_",
	)

	sanitized := replacer.Replace(component)

	// Remove leading/trailing underscores and dots
	sanitized = strings.Trim(sanitized, "_.")

	// Ensure we don't have an empty string
	if sanitized == "" {
		sanitized = "unknown"
	}

	return sanitized
}
