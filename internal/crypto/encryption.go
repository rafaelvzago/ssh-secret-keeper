package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// DefaultIterations for PBKDF2
	DefaultIterations = 100000
	// SaltSize in bytes
	SaltSize = 32
	// KeySize for AES-256
	KeySize = 32
)

// EncryptedData represents encrypted data with metadata
type EncryptedData struct {
	Data       string `json:"data"`        // Base64 encoded encrypted data
	Salt       string `json:"salt"`        // Base64 encoded salt
	IV         string `json:"iv"`          // Base64 encoded IV
	Algorithm  string `json:"algorithm"`   // Encryption algorithm
	Iterations int    `json:"iterations"`  // PBKDF2 iterations
}

// Encryptor handles encryption and decryption operations
type Encryptor struct {
	iterations int
}

// NewEncryptor creates a new encryptor with specified iterations
func NewEncryptor(iterations int) *Encryptor {
	if iterations <= 0 {
		iterations = DefaultIterations
	}
	
	return &Encryptor{
		iterations: iterations,
	}
}

// Encrypt encrypts data with the given passphrase using AES-256-GCM
func (e *Encryptor) Encrypt(data []byte, passphrase string) (*EncryptedData, error) {
	// Generate random salt
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(passphrase), salt, e.iterations, KeySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random IV
	iv := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt data
	ciphertext := gcm.Seal(nil, iv, data, nil)

	return &EncryptedData{
		Data:       base64.StdEncoding.EncodeToString(ciphertext),
		Salt:       base64.StdEncoding.EncodeToString(salt),
		IV:         base64.StdEncoding.EncodeToString(iv),
		Algorithm:  "AES-256-GCM",
		Iterations: e.iterations,
	}, nil
}

// Decrypt decrypts data with the given passphrase
func (e *Encryptor) Decrypt(encData *EncryptedData, passphrase string) ([]byte, error) {
	if encData.Algorithm != "AES-256-GCM" {
		return nil, fmt.Errorf("unsupported algorithm: %s", encData.Algorithm)
	}

	// Decode base64 data
	ciphertext, err := base64.StdEncoding.DecodeString(encData.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	salt, err := base64.StdEncoding.DecodeString(encData.Salt)
	if err != nil {
		return nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(encData.IV)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %w", err)
	}

	// Derive key using stored parameters
	key := pbkdf2.Key([]byte(passphrase), salt, encData.Iterations, KeySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt data
	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	// Clear sensitive data
	for i := range key {
		key[i] = 0
	}

	return plaintext, nil
}

// EncryptFiles encrypts multiple files with the same passphrase
func (e *Encryptor) EncryptFiles(files map[string][]byte, passphrase string) (map[string]*EncryptedData, error) {
	encrypted := make(map[string]*EncryptedData)
	
	for filename, data := range files {
		encData, err := e.Encrypt(data, passphrase)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt file %s: %w", filename, err)
		}
		encrypted[filename] = encData
	}
	
	return encrypted, nil
}

// DecryptFiles decrypts multiple files with the same passphrase
func (e *Encryptor) DecryptFiles(encrypted map[string]*EncryptedData, passphrase string) (map[string][]byte, error) {
	files := make(map[string][]byte)
	
	for filename, encData := range encrypted {
		data, err := e.Decrypt(encData, passphrase)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt file %s: %w", filename, err)
		}
		files[filename] = data
	}
	
	return files, nil
}

// GeneratePassphrase generates a random passphrase
func GeneratePassphrase(length int) (string, error) {
	if length < 16 {
		length = 32 // Default to 32 characters
	}
	
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// VerifyPassphrase verifies if a passphrase can decrypt the given data
func (e *Encryptor) VerifyPassphrase(encData *EncryptedData, passphrase string) bool {
	_, err := e.Decrypt(encData, passphrase)
	return err == nil
}
