package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rs/zerolog/log"
)

// DetectPlatform returns the current platform information
func DetectPlatform() Platform {
	return Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// GetBinaryName returns the expected binary name for the platform
func GetBinaryName(platform Platform) string {
	if platform.OS == "windows" {
		return "sshsk.exe"
	}
	return "sshsk"
}

// GetAssetName returns the expected asset name for a platform
func GetAssetName(platform Platform, version string) string {
	// Format: ssh-secret-keeper-{version}-{os}-{arch}.tar.gz
	// Example: ssh-secret-keeper-1.0.4-linux-amd64.tar.gz
	cleanVersion := strings.TrimPrefix(version, "v")
	return fmt.Sprintf("ssh-secret-keeper-%s-%s-%s.tar.gz",
		cleanVersion, platform.OS, platform.Arch)
}

// GetCurrentBinaryPath returns the path to the currently running binary
func GetCurrentBinaryPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve any symlinks
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		// If we can't resolve symlinks, use the original path
		log.Debug().Err(err).Msg("Could not resolve symlinks, using original path")
		return execPath, nil
	}

	return realPath, nil
}

// GetInstallDir returns the directory where the binary is installed
func GetInstallDir() (string, error) {
	binaryPath, err := GetCurrentBinaryPath()
	if err != nil {
		return "", err
	}
	return filepath.Dir(binaryPath), nil
}

// CanWriteToBinary checks if we have permission to update the binary
func CanWriteToBinary() error {
	binaryPath, err := GetCurrentBinaryPath()
	if err != nil {
		return fmt.Errorf("failed to get binary path: %w", err)
	}

	// Check if we can write to the binary location
	dir := filepath.Dir(binaryPath)
	testFile := filepath.Join(dir, ".sshsk-update-test")

	// Try to create a test file
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("insufficient permissions to update binary in %s: %w", dir, err)
	}
	file.Close()

	// Clean up test file
	os.Remove(testFile)

	return nil
}

// GetBackupPath returns the path for backing up the current binary
func GetBackupPath() (string, error) {
	binaryPath, err := GetCurrentBinaryPath()
	if err != nil {
		return "", err
	}
	return binaryPath + ".backup", nil
}

// IsSystemInstall checks if the binary is installed in a system directory
func IsSystemInstall() bool {
	binaryPath, err := GetCurrentBinaryPath()
	if err != nil {
		return false
	}

	systemDirs := []string{
		"/usr/local/bin",
		"/usr/bin",
		"/opt",
		"C:\\Program Files",
	}

	for _, dir := range systemDirs {
		if strings.HasPrefix(binaryPath, dir) {
			return true
		}
	}

	return false
}

// RequiresSudo checks if sudo/admin privileges might be required for update
func RequiresSudo() bool {
	if runtime.GOOS == "windows" {
		// On Windows, check if in Program Files
		binaryPath, err := GetCurrentBinaryPath()
		if err != nil {
			return false
		}
		return strings.Contains(binaryPath, "Program Files")
	}

	// On Unix-like systems, check if we can write
	err := CanWriteToBinary()
	return err != nil
}
