package analyzer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// Analyzer analyzes SSH directories and categorizes files
type Analyzer struct {
	detectors       []KeyDetector
	servicePatterns map[string][]string
	purposeRules    map[string]KeyPurpose
}

// New creates a new analyzer with default detectors
func New() *Analyzer {
	return &Analyzer{
		detectors: []KeyDetector{
			&RSAKeyDetector{},
			&PEMKeyDetector{},
			&OpenSSHKeyDetector{},
			&ConfigFileDetector{},
			&KnownHostsDetector{},
			&AuthorizedKeysDetector{},
			&UniversalFileDetector{}, // Must be last to catch all undetected files
		},
		servicePatterns: map[string][]string{
			"github":    {"*github*"},
			"gitlab":    {"*gitlab*"},
			"bitbucket": {"*bitbucket*"},
			"argocd":    {"*argocd*"},
			"quay":      {"*quay*"},
			"gke":       {"*gke*", "*gcp*", "*google*"},
			"aws":       {"*aws*", "*ec2*"},
		},
		purposeRules: map[string]KeyPurpose{
			"*work*":     PurposeWork,
			"*corp*":     PurposeWork,
			"*office*":   PurposeWork,
			"*personal*": PurposePersonal,
			"id_rsa":     PurposePersonal,
		},
	}
}

// AnalyzeDirectory analyzes an SSH directory and returns categorized results
func (a *Analyzer) AnalyzeDirectory(sshDir string) (*DetectionResult, error) {
	log.Info().Str("dir", sshDir).Msg("Starting SSH directory analysis")

	// Read directory
	files, err := os.ReadDir(sshDir)
	if err != nil {
		return nil, fmt.Errorf("error reading SSH directory: %w", err)
	}

	var keys []KeyInfo
	allFilenames := make([]string, 0, len(files))

	// First pass: collect all filenames
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		allFilenames = append(allFilenames, file.Name())
	}

	// Second pass: analyze each file
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(sshDir, file.Name())
		keyInfo, err := a.analyzeFile(filePath, file, allFilenames)
		if err != nil {
			log.Warn().Err(err).Str("file", file.Name()).Msg("Failed to analyze file")
			continue
		}

		if keyInfo != nil {
			keys = append(keys, *keyInfo)
		}
	}

	// Post-process: find key pairs and categorize
	result := a.processResults(keys)

	log.Info().
		Int("total_files", result.Summary.TotalFiles).
		Int("key_pairs", result.Summary.KeyPairCount).
		Int("services", result.Summary.ServiceKeys).
		Msg("Analysis completed")

	return result, nil
}

// analyzeFile analyzes a single file
func (a *Analyzer) analyzeFile(filePath string, fileInfo fs.DirEntry, allFiles []string) (*KeyInfo, error) {
	// Get file info
	info, err := fileInfo.Info()
	if err != nil {
		return nil, fmt.Errorf("error getting file info: %w", err)
	}

	// Read file content (first 4KB should be enough for detection)
	content, err := a.readFileHead(filePath, 4096)
	if err != nil {
		return nil, fmt.Errorf("error reading file content: %w", err)
	}

	// Try each detector
	for _, detector := range a.detectors {
		if keyInfo, detected := detector.Detect(fileInfo.Name(), content); detected {
			// Fill in additional info
			keyInfo.Filename = fileInfo.Name()
			keyInfo.Permissions = info.Mode()
			keyInfo.Size = info.Size()
			keyInfo.ModTime = info.ModTime()

			// Determine service and purpose
			a.enhanceKeyInfo(keyInfo, allFiles)

			log.Debug().
				Str("file", fileInfo.Name()).
				Str("detector", detector.Name()).
				Str("type", string(keyInfo.Type)).
				Msg("File detected")

			return keyInfo, nil
		}
	}

	// No detector matched
	log.Debug().Str("file", fileInfo.Name()).Msg("File not recognized")
	return nil, nil
}

// readFileHead reads the first n bytes of a file
func (a *Analyzer) readFileHead(filePath string, maxBytes int) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buffer := make([]byte, maxBytes)
	n, err := file.Read(buffer)
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}

	return buffer[:n], nil
}

// enhanceKeyInfo adds service and purpose information to key info
func (a *Analyzer) enhanceKeyInfo(keyInfo *KeyInfo, allFiles []string) {
	filename := strings.ToLower(keyInfo.Filename)

	// Determine service
	for service, patterns := range a.servicePatterns {
		for _, pattern := range patterns {
			if matched, _ := filepath.Match(strings.ToLower(pattern), filename); matched {
				keyInfo.Service = service
				keyInfo.Purpose = PurposeService
				return
			}
		}
	}

	// Determine purpose from rules
	for pattern, purpose := range a.purposeRules {
		if matched, _ := filepath.Match(strings.ToLower(pattern), filename); matched {
			keyInfo.Purpose = purpose
			return
		}
	}

	// Default purpose based on type
	switch keyInfo.Type {
	case KeyTypeConfig, KeyTypeHosts, KeyTypeAuthorized:
		keyInfo.Purpose = PurposeSystem
	default:
		if keyInfo.Service != "" {
			keyInfo.Purpose = PurposeService
		} else {
			keyInfo.Purpose = PurposePersonal
		}
	}
}

// processResults processes the analyzed keys to find pairs and categorize
func (a *Analyzer) processResults(keys []KeyInfo) *DetectionResult {
	keyPairs := a.findKeyPairs(keys)
	categories := a.categorizeKeys(keys)

	var systemFiles, unknownFiles []KeyInfo
	for _, key := range keys {
		if key.Type == KeyTypeConfig || key.Type == KeyTypeHosts || key.Type == KeyTypeAuthorized {
			systemFiles = append(systemFiles, key)
		} else if key.Type == KeyTypeUnknown {
			unknownFiles = append(unknownFiles, key)
		}
	}

	summary := a.generateSummary(keys, keyPairs)

	return &DetectionResult{
		Keys:         keys,
		KeyPairs:     keyPairs,
		Categories:   categories,
		SystemFiles:  systemFiles,
		UnknownFiles: unknownFiles,
		Summary:      summary,
	}
}

// findKeyPairs identifies related key files
func (a *Analyzer) findKeyPairs(keys []KeyInfo) map[string]*KeyPairInfo {
	pairs := make(map[string]*KeyPairInfo)

	for _, key := range keys {
		if key.Type != KeyTypePrivate && key.Type != KeyTypePublic {
			continue
		}

		baseName := a.getBaseName(key.Filename)

		if pair, exists := pairs[baseName]; exists {
			if key.Type == KeyTypePrivate {
				pair.PrivateKeyFile = key.Filename
			} else {
				pair.PublicKeyFile = key.Filename
			}
		} else {
			pair := &KeyPairInfo{
				BaseName: baseName,
			}
			if key.Type == KeyTypePrivate {
				pair.PrivateKeyFile = key.Filename
			} else {
				pair.PublicKeyFile = key.Filename
			}
			pairs[baseName] = pair
		}
	}

	return pairs
}

// getBaseName extracts the base name from a key filename
func (a *Analyzer) getBaseName(filename string) string {
	// Remove common extensions
	name := strings.TrimSuffix(filename, ".pub")
	name = strings.TrimSuffix(name, ".pem")
	name = strings.TrimSuffix(name, ".rsa")
	name = strings.TrimSuffix(name, ".dsa")
	name = strings.TrimSuffix(name, ".ecdsa")
	name = strings.TrimSuffix(name, ".ed25519")

	return name
}

// categorizeKeys groups keys by purpose
func (a *Analyzer) categorizeKeys(keys []KeyInfo) map[string][]KeyInfo {
	categories := make(map[string][]KeyInfo)

	for _, key := range keys {
		category := string(key.Purpose)
		categories[category] = append(categories[category], key)
	}

	return categories
}

// generateSummary creates a summary of the analysis
func (a *Analyzer) generateSummary(keys []KeyInfo, keyPairs map[string]*KeyPairInfo) *AnalysisSummary {
	summary := &AnalysisSummary{
		TotalFiles:       len(keys),
		KeyPairCount:     len(keyPairs),
		FormatBreakdown:  make(map[KeyFormat]int),
		PurposeBreakdown: make(map[KeyPurpose]int),
	}

	for _, key := range keys {
		summary.FormatBreakdown[key.Format]++
		summary.PurposeBreakdown[key.Purpose]++

		switch key.Purpose {
		case PurposeService:
			summary.ServiceKeys++
		case PurposePersonal:
			summary.PersonalKeys++
		case PurposeWork:
			summary.WorkKeys++
		case PurposeSystem:
			summary.SystemFiles++
		}

		if key.Type == KeyTypeUnknown {
			summary.UnknownFiles++
		}
	}

	return summary
}
