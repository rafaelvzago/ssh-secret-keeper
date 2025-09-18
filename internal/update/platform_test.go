package update

import (
	"runtime"
	"strings"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	platform := DetectPlatform()

	if platform.OS != runtime.GOOS {
		t.Errorf("DetectPlatform().OS = %s, want %s", platform.OS, runtime.GOOS)
	}

	if platform.Arch != runtime.GOARCH {
		t.Errorf("DetectPlatform().Arch = %s, want %s", platform.Arch, runtime.GOARCH)
	}
}

func TestGetBinaryName(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		expected string
	}{
		{
			"Linux binary",
			Platform{OS: "linux", Arch: "amd64"},
			"sshsk",
		},
		{
			"macOS binary",
			Platform{OS: "darwin", Arch: "arm64"},
			"sshsk",
		},
		{
			"Windows binary",
			Platform{OS: "windows", Arch: "amd64"},
			"sshsk.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBinaryName(tt.platform)
			if result != tt.expected {
				t.Errorf("GetBinaryName(%v) = %s, want %s",
					tt.platform, result, tt.expected)
			}
		})
	}
}

func TestGetAssetName(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		version  string
		expected string
	}{
		{
			"Linux amd64",
			Platform{OS: "linux", Arch: "amd64"},
			"v1.0.4",
			"ssh-secret-keeper-1.0.4-linux-amd64.tar.gz",
		},
		{
			"macOS arm64",
			Platform{OS: "darwin", Arch: "arm64"},
			"v1.0.4",
			"ssh-secret-keeper-1.0.4-darwin-arm64.tar.gz",
		},
		{
			"Windows amd64",
			Platform{OS: "windows", Arch: "amd64"},
			"1.0.4",
			"ssh-secret-keeper-1.0.4-windows-amd64.tar.gz",
		},
		{
			"Linux arm64",
			Platform{OS: "linux", Arch: "arm64"},
			"v2.0.0",
			"ssh-secret-keeper-2.0.0-linux-arm64.tar.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAssetName(tt.platform, tt.version)
			if result != tt.expected {
				t.Errorf("GetAssetName(%v, %s) = %s, want %s",
					tt.platform, tt.version, result, tt.expected)
			}
		})
	}
}

func TestGetBackupPath(t *testing.T) {
	path, err := GetBackupPath()
	if err != nil {
		// This might fail in test environment if binary path can't be determined
		t.Skipf("GetBackupPath() failed: %v", err)
	}

	if !strings.HasSuffix(path, ".backup") {
		t.Errorf("GetBackupPath() = %s, expected to end with .backup", path)
	}
}

func TestIsSystemInstall(t *testing.T) {
	// This test is environment-dependent
	result := IsSystemInstall()
	t.Logf("IsSystemInstall() = %v (environment-dependent)", result)
}

func TestRequiresSudo(t *testing.T) {
	// This test is environment-dependent
	result := RequiresSudo()
	t.Logf("RequiresSudo() = %v (environment-dependent)", result)
}
