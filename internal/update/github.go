package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	defaultGitHubRepo = "rafaelvzago/ssh-secret-keeper"
	githubAPIBase     = "https://api.github.com"
	githubTimeout     = 30 * time.Second
)

// GitHubClient handles communication with GitHub API
type GitHubClient struct {
	repo       string
	httpClient *http.Client
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient(repo string) *GitHubClient {
	if repo == "" {
		repo = defaultGitHubRepo
	}

	return &GitHubClient{
		repo: repo,
		httpClient: &http.Client{
			Timeout: githubTimeout,
		},
	}
}

// GetLatestRelease fetches the latest release from GitHub
func (c *GitHubClient) GetLatestRelease(includePreRelease bool) (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", githubAPIBase, c.repo)

	if includePreRelease {
		// If including pre-releases, we need to get all releases and find the latest
		return c.getLatestFromAllReleases()
	}

	log.Debug().Str("url", url).Msg("Fetching latest release from GitHub")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ssh-secret-keeper-updater")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release: %w", err)
	}

	return &release, nil
}

// GetReleaseByTag fetches a specific release by tag
func (c *GitHubClient) GetReleaseByTag(tag string) (*Release, error) {
	// Clean the tag
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}

	url := fmt.Sprintf("%s/repos/%s/releases/tags/%s", githubAPIBase, c.repo, tag)

	log.Debug().Str("url", url).Str("tag", tag).Msg("Fetching release by tag from GitHub")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ssh-secret-keeper-updater")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release %s not found", tag)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release: %w", err)
	}

	return &release, nil
}

// getLatestFromAllReleases gets the latest release including pre-releases
func (c *GitHubClient) getLatestFromAllReleases() (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/releases", githubAPIBase, c.repo)

	log.Debug().Str("url", url).Msg("Fetching all releases from GitHub")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ssh-secret-keeper-updater")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode releases: %w", err)
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found")
	}

	// Return the first one (most recent)
	return &releases[0], nil
}

// FindAssetForPlatform finds the appropriate asset for the given platform
func (c *GitHubClient) FindAssetForPlatform(release *Release, platform Platform) (*Asset, error) {
	expectedName := GetAssetName(platform, release.TagName)

	log.Debug().
		Str("platform", fmt.Sprintf("%s-%s", platform.OS, platform.Arch)).
		Str("expected_asset", expectedName).
		Int("total_assets", len(release.Assets)).
		Msg("Looking for platform-specific asset")

	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			return &asset, nil
		}
	}

	// List available assets for debugging
	var available []string
	for _, asset := range release.Assets {
		available = append(available, asset.Name)
	}

	return nil, fmt.Errorf("no asset found for platform %s-%s (expected: %s, available: %s)",
		platform.OS, platform.Arch, expectedName, strings.Join(available, ", "))
}

// GetChecksumForAsset fetches the checksum for a specific asset
func (c *GitHubClient) GetChecksumForAsset(release *Release, assetName string) (string, error) {
	// Look for checksums.txt in the assets
	var checksumAsset *Asset
	for _, asset := range release.Assets {
		if asset.Name == "checksums.txt" {
			checksumAsset = &asset
			break
		}
	}

	if checksumAsset == nil {
		log.Warn().Msg("No checksums.txt file found in release assets")
		return "", nil // Return empty string if no checksum file
	}

	// Download the checksum file
	resp, err := c.httpClient.Get(checksumAsset.BrowserDownloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download checksums (status %d)", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read checksums: %w", err)
	}

	// Parse checksums file (format: "checksum  filename")
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == assetName {
			return parts[0], nil
		}
	}

	log.Warn().Str("asset", assetName).Msg("Checksum not found for asset")
	return "", nil
}
