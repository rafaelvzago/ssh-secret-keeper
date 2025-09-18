package update

import (
	"time"
)

// Release represents a GitHub release
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"` // Release notes
	Assets      []Asset   `json:"assets"`
}

// Asset represents a release asset (downloadable file)
type Asset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
}

// UpdateStatus represents the current update status
type UpdateStatus struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	ReleaseNotes    string
	PublishedAt     time.Time
}

// Platform represents the current system platform
type Platform struct {
	OS   string // darwin, linux, windows
	Arch string // amd64, arm64
}

// UpdateOptions configures the update behavior
type UpdateOptions struct {
	Version      string // Specific version to update to (empty for latest)
	CheckOnly    bool   // Only check for updates, don't install
	Force        bool   // Force update even if already on latest
	PreRelease   bool   // Include pre-release versions
	NoBackup     bool   // Don't create backup of current binary
	SkipChecksum bool   // Skip checksum verification (not recommended)
	SkipVerify   bool   // Skip new binary verification
}

// UpdateConfig represents update configuration
type UpdateConfig struct {
	CheckOnStartup bool          `yaml:"check_on_startup" mapstructure:"check_on_startup"`
	AutoUpdate     bool          `yaml:"auto_update" mapstructure:"auto_update"`
	Channel        string        `yaml:"channel" mapstructure:"channel"`
	CheckInterval  time.Duration `yaml:"check_interval" mapstructure:"check_interval"`
	GitHubRepo     string        `yaml:"github_repo" mapstructure:"github_repo"`
}
