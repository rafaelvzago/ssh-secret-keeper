package analyzer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// Service provides SSH file analysis functionality
type Service struct {
	detectors       []KeyDetector
	servicePatterns map[string][]string
	purposeRules    map[string]KeyPurpose
}

// NewService creates a new analysis service with default configuration
func NewService() *Service {
	service := &Service{
		detectors:       make([]KeyDetector, 0),
		servicePatterns: getDefaultServicePatterns(),
		purposeRules:    getDefaultPurposeRules(),
	}

	// Register default detectors
	service.registerDefaultDetectors()

	return service
}

// NewServiceWithConfig creates a new analysis service with custom configuration
func NewServiceWithConfig(servicePatterns map[string][]string, purposeRules map[string]KeyPurpose) *Service {
	service := &Service{
		detectors:       make([]KeyDetector, 0),
		servicePatterns: servicePatterns,
		purposeRules:    purposeRules,
	}

	// Register default detectors
	service.registerDefaultDetectors()

	return service
}

// RegisterDetector registers a new key detector (Open/Closed Principle)
func (s *Service) RegisterDetector(detector KeyDetector) {
	if detector == nil {
		log.Warn().Msg("Attempted to register nil detector")
		return
	}

	// Check if detector already exists
	for _, existing := range s.detectors {
		if existing.Name() == detector.Name() {
			log.Warn().Str("detector", detector.Name()).Msg("Detector already registered, skipping")
			return
		}
	}

	s.detectors = append(s.detectors, detector)
	log.Debug().Str("detector", detector.Name()).Msg("Detector registered")
}

// GetRegisteredDetectors returns all registered detectors
func (s *Service) GetRegisteredDetectors() []KeyDetector {
	// Return a copy to prevent external modification
	detectors := make([]KeyDetector, len(s.detectors))
	copy(detectors, s.detectors)
	return detectors
}

// AnalyzeDirectory analyzes an SSH directory and returns categorized results
func (s *Service) AnalyzeDirectory(sshDir string) (*DetectionResult, error) {
	log.Info().Str("dir", sshDir).Msg("Starting SSH directory analysis")

	// Validate directory
	if err := s.validateDirectory(sshDir); err != nil {
		return nil, fmt.Errorf("directory validation failed: %w", err)
	}

	// Read directory
	files, err := os.ReadDir(sshDir)
	if err != nil {
		return nil, fmt.Errorf("error reading SSH directory: %w", err)
	}

	var keys []KeyInfo
	allFilenames := s.extractFilenames(files)

	// Analyze each file
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(sshDir, file.Name())
		keyInfo, err := s.AnalyzeFile(filePath)
		if err != nil {
			log.Warn().Err(err).Str("file", file.Name()).Msg("Failed to analyze file")
			continue
		}

		if keyInfo != nil {
			// Enhance with context from all files
			s.enhanceKeyInfo(keyInfo, allFilenames)
			keys = append(keys, *keyInfo)
		}
	}

	// Process results
	result := s.processAnalysisResults(keys)

	log.Info().
		Int("total_files", result.Summary.TotalFiles).
		Int("key_pairs", result.Summary.KeyPairCount).
		Int("services", result.Summary.ServiceKeys).
		Msg("Analysis completed")

	return result, nil
}

// AnalyzeFile analyzes a single file
func (s *Service) AnalyzeFile(filePath string) (*KeyInfo, error) {
	// Get file info
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot stat file: %w", err)
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file")
	}

	// Read file content (first 4KB should be enough for detection)
	content, err := s.readFileHead(filePath, 4096)
	if err != nil {
		return nil, fmt.Errorf("error reading file content: %w", err)
	}

	filename := filepath.Base(filePath)

	// Try each registered detector
	for _, detector := range s.detectors {
		if keyInfo, detected := detector.Detect(filename, content); detected {
			// Fill in file system information
			keyInfo.Filename = filename
			keyInfo.Permissions = stat.Mode()
			keyInfo.Size = stat.Size()
			keyInfo.ModTime = stat.ModTime()

			log.Debug().
				Str("file", filename).
				Str("detector", detector.Name()).
				Str("type", string(keyInfo.Type)).
				Msg("File detected")

			return keyInfo, nil
		}
	}

	// No detector matched
	log.Debug().Str("file", filename).Msg("File not recognized by any detector")
	return nil, nil
}

// Private helper methods

func (s *Service) registerDefaultDetectors() {
	defaultDetectors := []KeyDetector{
		&RSAKeyDetector{},
		&PEMKeyDetector{},
		&OpenSSHKeyDetector{},
		&ConfigFileDetector{},
		&KnownHostsDetector{},
		&AuthorizedKeysDetector{},
	}

	for _, detector := range defaultDetectors {
		s.RegisterDetector(detector)
	}
}

func (s *Service) validateDirectory(sshDir string) error {
	stat, err := os.Stat(sshDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", sshDir)
		}
		return fmt.Errorf("cannot access directory: %w", err)
	}

	if !stat.IsDir() {
		return fmt.Errorf("path is not a directory: %s", sshDir)
	}

	return nil
}

func (s *Service) extractFilenames(files []fs.DirEntry) []string {
	filenames := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			filenames = append(filenames, file.Name())
		}
	}
	return filenames
}

func (s *Service) readFileHead(filePath string, maxBytes int) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, maxBytes)
	n, err := file.Read(buffer)
	if err != nil && err.Error() != "EOF" {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	return buffer[:n], nil
}

func (s *Service) enhanceKeyInfo(keyInfo *KeyInfo, allFiles []string) {
	filename := filepath.Base(keyInfo.Filename)

	// Determine service from patterns
	if service := s.detectService(filename); service != "" {
		keyInfo.Service = service
		keyInfo.Purpose = PurposeService
		return
	}

	// Determine purpose from rules
	if purpose := s.detectPurpose(filename); purpose != "" {
		keyInfo.Purpose = KeyPurpose(purpose)
		return
	}

	// Default purpose based on type
	keyInfo.Purpose = s.getDefaultPurpose(keyInfo.Type, keyInfo.Service)
}

func (s *Service) detectService(filename string) string {
	filename = filepath.Base(filename)
	for service, patterns := range s.servicePatterns {
		for _, pattern := range patterns {
			if matched, _ := filepath.Match(pattern, filename); matched {
				return service
			}
		}
	}
	return ""
}

func (s *Service) detectPurpose(filename string) string {
	filename = filepath.Base(filename)
	for pattern, purpose := range s.purposeRules {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return string(purpose)
		}
	}
	return ""
}

func (s *Service) getDefaultPurpose(keyType KeyType, service string) KeyPurpose {
	switch keyType {
	case KeyTypeConfig, KeyTypeHosts, KeyTypeAuthorized:
		return PurposeSystem
	default:
		if service != "" {
			return PurposeService
		}
		return PurposePersonal
	}
}

func (s *Service) processAnalysisResults(keys []KeyInfo) *DetectionResult {
	keyPairs := s.findKeyPairs(keys)
	categories := s.categorizeKeys(keys)
	summary := s.generateSummary(keys, keyPairs)

	var systemFiles, unknownFiles []KeyInfo
	for _, key := range keys {
		switch key.Type {
		case KeyTypeConfig, KeyTypeHosts, KeyTypeAuthorized:
			systemFiles = append(systemFiles, key)
		case KeyTypeUnknown:
			unknownFiles = append(unknownFiles, key)
		}
	}

	return &DetectionResult{
		Keys:         keys,
		KeyPairs:     keyPairs,
		Categories:   categories,
		SystemFiles:  systemFiles,
		UnknownFiles: unknownFiles,
		Summary:      summary,
	}
}

func (s *Service) findKeyPairs(keys []KeyInfo) map[string]*KeyPairInfo {
	pairs := make(map[string]*KeyPairInfo)

	for _, key := range keys {
		if key.Type != KeyTypePrivate && key.Type != KeyTypePublic {
			continue
		}

		baseName := s.getBaseName(key.Filename)

		if pair, exists := pairs[baseName]; exists {
			s.updateKeyPair(pair, key)
		} else {
			pairs[baseName] = s.createKeyPair(baseName, key)
		}
	}

	return pairs
}

func (s *Service) getBaseName(filename string) string {
	// Remove common extensions
	extensions := []string{".pub", ".pem", ".rsa", ".dsa", ".ecdsa", ".ed25519"}

	baseName := filename
	for _, ext := range extensions {
		if len(baseName) > len(ext) && baseName[len(baseName)-len(ext):] == ext {
			baseName = baseName[:len(baseName)-len(ext)]
			break
		}
	}

	return baseName
}

func (s *Service) createKeyPair(baseName string, key KeyInfo) *KeyPairInfo {
	pair := &KeyPairInfo{BaseName: baseName}
	s.updateKeyPair(pair, key)
	return pair
}

func (s *Service) updateKeyPair(pair *KeyPairInfo, key KeyInfo) {
	if key.Type == KeyTypePrivate {
		pair.PrivateKeyFile = key.Filename
	} else if key.Type == KeyTypePublic {
		pair.PublicKeyFile = key.Filename
	}
}

func (s *Service) categorizeKeys(keys []KeyInfo) map[string][]KeyInfo {
	categories := make(map[string][]KeyInfo)

	for _, key := range keys {
		category := string(key.Purpose)
		categories[category] = append(categories[category], key)
	}

	return categories
}

func (s *Service) generateSummary(keys []KeyInfo, keyPairs map[string]*KeyPairInfo) *AnalysisSummary {
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
