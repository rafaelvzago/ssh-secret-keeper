package crypto

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

// Service provides encryption functionality with enhanced error handling and validation
type Service struct {
	encryptor *Encryptor
}

// NewService creates a new encryption service with default settings
func NewService() *Service {
	return &Service{
		encryptor: NewEncryptor(DefaultIterations),
	}
}

// NewServiceWithIterations creates a new encryption service with custom iterations
func NewServiceWithIterations(iterations int) *Service {
	if iterations < 10000 {
		log.Warn().
			Int("requested", iterations).
			Int("minimum", 10000).
			Msg("Iteration count too low, using minimum recommended value")
		iterations = 10000
	}

	return &Service{
		encryptor: NewEncryptor(iterations),
	}
}

// Encrypt encrypts data with the given passphrase using AES-256-GCM
func (s *Service) Encrypt(data []byte, passphrase string) (*EncryptedData, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("cannot encrypt empty data")
	}

	if err := s.validatePassphrase(passphrase); err != nil {
		return nil, fmt.Errorf("passphrase validation failed: %w", err)
	}

	encrypted, err := s.encryptor.Encrypt(data, passphrase)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	log.Debug().
		Int("data_size", len(data)).
		Str("algorithm", encrypted.Algorithm).
		Int("iterations", encrypted.Iterations).
		Msg("Data encrypted successfully")

	return encrypted, nil
}

// Decrypt decrypts data with the given passphrase
func (s *Service) Decrypt(encData *EncryptedData, passphrase string) ([]byte, error) {
	if encData == nil {
		return nil, fmt.Errorf("encrypted data is nil")
	}

	if err := s.validateEncryptedData(encData); err != nil {
		return nil, fmt.Errorf("encrypted data validation failed: %w", err)
	}

	if err := s.validatePassphrase(passphrase); err != nil {
		return nil, fmt.Errorf("passphrase validation failed: %w", err)
	}

	decrypted, err := s.encryptor.Decrypt(encData, passphrase)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	log.Debug().
		Int("decrypted_size", len(decrypted)).
		Str("algorithm", encData.Algorithm).
		Msg("Data decrypted successfully")

	return decrypted, nil
}

// EncryptFiles encrypts multiple files with the same passphrase
func (s *Service) EncryptFiles(files map[string][]byte, passphrase string) (map[string]*EncryptedData, error) {
	if len(files) == 0 {
		return make(map[string]*EncryptedData), nil
	}

	if err := s.validatePassphrase(passphrase); err != nil {
		return nil, fmt.Errorf("passphrase validation failed: %w", err)
	}

	encrypted, err := s.encryptor.EncryptFiles(files, passphrase)
	if err != nil {
		return nil, fmt.Errorf("batch encryption failed: %w", err)
	}

	log.Info().
		Int("files_encrypted", len(encrypted)).
		Msg("Batch encryption completed")

	return encrypted, nil
}

// DecryptFiles decrypts multiple files with the same passphrase
func (s *Service) DecryptFiles(encrypted map[string]*EncryptedData, passphrase string) (map[string][]byte, error) {
	if len(encrypted) == 0 {
		return make(map[string][]byte), nil
	}

	if err := s.validatePassphrase(passphrase); err != nil {
		return nil, fmt.Errorf("passphrase validation failed: %w", err)
	}

	// Validate all encrypted data first
	for filename, encData := range encrypted {
		if err := s.validateEncryptedData(encData); err != nil {
			return nil, fmt.Errorf("encrypted data validation failed for file %s: %w", filename, err)
		}
	}

	decrypted, err := s.encryptor.DecryptFiles(encrypted, passphrase)
	if err != nil {
		return nil, fmt.Errorf("batch decryption failed: %w", err)
	}

	log.Info().
		Int("files_decrypted", len(decrypted)).
		Msg("Batch decryption completed")

	return decrypted, nil
}

// VerifyPassphrase verifies if a passphrase can decrypt the given data
func (s *Service) VerifyPassphrase(encData *EncryptedData, passphrase string) bool {
	if encData == nil || passphrase == "" {
		return false
	}

	if err := s.validateEncryptedData(encData); err != nil {
		log.Debug().Err(err).Msg("Encrypted data validation failed during passphrase verification")
		return false
	}

	result := s.encryptor.VerifyPassphrase(encData, passphrase)

	log.Debug().
		Bool("verification_result", result).
		Msg("Passphrase verification completed")

	return result
}

// GeneratePassphrase generates a cryptographically secure passphrase
func (s *Service) GeneratePassphrase(length int) (string, error) {
	const minLength = 16
	const maxLength = 512

	if length < minLength {
		log.Warn().
			Int("requested", length).
			Int("minimum", minLength).
			Msg("Passphrase length too short, using minimum")
		length = minLength
	}

	if length > maxLength {
		log.Warn().
			Int("requested", length).
			Int("maximum", maxLength).
			Msg("Passphrase length too long, using maximum")
		length = maxLength
	}

	passphrase, err := GeneratePassphrase(length)
	if err != nil {
		return "", fmt.Errorf("passphrase generation failed: %w", err)
	}

	log.Info().
		Int("length", len(passphrase)).
		Msg("Secure passphrase generated")

	return passphrase, nil
}

// Private validation methods

func (s *Service) validatePassphrase(passphrase string) error {
	if passphrase == "" {
		return fmt.Errorf("passphrase cannot be empty")
	}

	const minLength = 8
	if len(passphrase) < minLength {
		return fmt.Errorf("passphrase too short (minimum %d characters)", minLength)
	}

	// Check for null bytes which could cause issues
	if strings.Contains(passphrase, "\x00") {
		return fmt.Errorf("passphrase contains null bytes")
	}

	return nil
}

func (s *Service) validateEncryptedData(encData *EncryptedData) error {
	if encData.Data == "" {
		return fmt.Errorf("encrypted data is empty")
	}

	if encData.Salt == "" {
		return fmt.Errorf("salt is empty")
	}

	if encData.IV == "" {
		return fmt.Errorf("IV is empty")
	}

	if encData.Algorithm == "" {
		return fmt.Errorf("algorithm is not specified")
	}

	// Verify supported algorithm
	supportedAlgorithms := []string{"AES-256-GCM"}
	algorithmSupported := false
	for _, algo := range supportedAlgorithms {
		if encData.Algorithm == algo {
			algorithmSupported = true
			break
		}
	}

	if !algorithmSupported {
		return fmt.Errorf("unsupported algorithm: %s", encData.Algorithm)
	}

	if encData.Iterations < 10000 {
		return fmt.Errorf("iteration count too low: %d (minimum 10000)", encData.Iterations)
	}

	if encData.Iterations > 1000000 {
		return fmt.Errorf("iteration count too high: %d (maximum 1000000)", encData.Iterations)
	}

	return nil
}

// GetEncryptionStats returns statistics about encryption operations
func (s *Service) GetEncryptionStats(encrypted map[string]*EncryptedData) map[string]interface{} {
	stats := make(map[string]interface{})

	if len(encrypted) == 0 {
		return stats
	}

	algorithmCounts := make(map[string]int)
	iterationCounts := make(map[int]int)
	var totalDataSize int

	for _, encData := range encrypted {
		algorithmCounts[encData.Algorithm]++
		iterationCounts[encData.Iterations]++
		totalDataSize += len(encData.Data)
	}

	stats["file_count"] = len(encrypted)
	stats["total_encrypted_size"] = totalDataSize
	stats["average_encrypted_size"] = float64(totalDataSize) / float64(len(encrypted))
	stats["algorithms"] = algorithmCounts
	stats["iteration_counts"] = iterationCounts

	return stats
}

// SecureWipe attempts to securely clear sensitive data from memory
func (s *Service) SecureWipe(data []byte) {
	if len(data) == 0 {
		return
	}

	// Overwrite with random data multiple times
	for i := 0; i < len(data); i++ {
		data[i] = 0
	}

	// Additional passes with different patterns
	for i := 0; i < len(data); i++ {
		data[i] = 0xFF
	}

	for i := 0; i < len(data); i++ {
		data[i] = 0xAA
	}

	for i := 0; i < len(data); i++ {
		data[i] = 0x55
	}

	// Final zero pass
	for i := 0; i < len(data); i++ {
		data[i] = 0
	}
}
