package update

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Downloader handles downloading and extracting release assets
type Downloader struct {
	httpClient *http.Client
	tempDir    string
}

// NewDownloader creates a new downloader
func NewDownloader() *Downloader {
	return &Downloader{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Longer timeout for downloads
		},
		tempDir: os.TempDir(),
	}
}

// DownloadAsset downloads an asset to a temporary file
func (d *Downloader) DownloadAsset(asset *Asset, progress func(downloaded, total int64)) (string, error) {
	// Create temporary file
	tempFile, err := os.CreateTemp(d.tempDir, "sshsk-update-*.tar.gz")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	tempPath := tempFile.Name()

	log.Info().
		Str("url", asset.BrowserDownloadURL).
		Str("file", filepath.Base(tempPath)).
		Int64("size", asset.Size).
		Msg("Downloading update")

	// Download the file
	resp, err := d.httpClient.Get(asset.BrowserDownloadURL)
	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(tempPath)
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create progress reader if callback provided
	var reader io.Reader = resp.Body
	if progress != nil {
		reader = &progressReader{
			reader:   resp.Body,
			total:    asset.Size,
			callback: progress,
		}
	}

	// Copy to temp file
	written, err := io.Copy(tempFile, reader)
	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to write download: %w", err)
	}

	log.Debug().
		Int64("bytes", written).
		Str("file", tempPath).
		Msg("Download completed")

	return tempPath, nil
}

// ExtractBinary extracts the binary from a tar.gz archive
func (d *Downloader) ExtractBinary(archivePath string, platform Platform) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Look for the binary
	binaryName := GetBinaryName(platform)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar: %w", err)
		}

		// Check if this is our binary
		if filepath.Base(header.Name) == binaryName {
			// Create temp file for binary
			tempBinary, err := os.CreateTemp(d.tempDir, "sshsk-new-*")
			if err != nil {
				return "", fmt.Errorf("failed to create temp binary: %w", err)
			}

			tempPath := tempBinary.Name()

			// Copy binary content
			_, err = io.Copy(tempBinary, tarReader)
			tempBinary.Close()

			if err != nil {
				os.Remove(tempPath)
				return "", fmt.Errorf("failed to extract binary: %w", err)
			}

			// Set executable permissions
			if err := os.Chmod(tempPath, 0755); err != nil {
				os.Remove(tempPath)
				return "", fmt.Errorf("failed to set permissions: %w", err)
			}

			log.Debug().
				Str("binary", binaryName).
				Str("extracted_to", tempPath).
				Msg("Binary extracted successfully")

			return tempPath, nil
		}
	}

	return "", fmt.Errorf("binary %s not found in archive", binaryName)
}

// VerifyChecksum verifies the SHA256 checksum of a file
func (d *Downloader) VerifyChecksum(filePath, expectedChecksum string) error {
	if expectedChecksum == "" {
		log.Warn().Msg("No checksum provided, skipping verification")
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualChecksum := fmt.Sprintf("%x", hash.Sum(nil))

	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	log.Debug().
		Str("checksum", actualChecksum[:16]+"...").
		Msg("Checksum verified successfully")

	return nil
}

// CleanupTempFiles removes temporary files created during download
func (d *Downloader) CleanupTempFiles() {
	// Clean up any sshsk-update-* or sshsk-new-* files in temp dir
	pattern := filepath.Join(d.tempDir, "sshsk-*")
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to list temp files for cleanup")
		return
	}

	for _, file := range files {
		if strings.Contains(file, "sshsk-update-") || strings.Contains(file, "sshsk-new-") {
			if err := os.Remove(file); err != nil {
				log.Debug().Err(err).Str("file", file).Msg("Failed to remove temp file")
			} else {
				log.Debug().Str("file", file).Msg("Cleaned up temp file")
			}
		}
	}
}

// progressReader wraps an io.Reader and reports progress
type progressReader struct {
	reader     io.Reader
	total      int64
	downloaded int64
	callback   func(downloaded, total int64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.downloaded += int64(n)

	if pr.callback != nil {
		pr.callback(pr.downloaded, pr.total)
	}

	return n, err
}
