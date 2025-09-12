package crypto

import (
	"testing"
)

func TestEncryptor_Encrypt(t *testing.T) {
	encryptor := NewEncryptor(1000) // Use lower iterations for faster tests
	
	testData := []byte("test ssh key data")
	passphrase := "test-passphrase"
	
	encrypted, err := encryptor.Encrypt(testData, passphrase)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}
	
	// Verify encrypted data structure
	if encrypted.Algorithm != "AES-256-GCM" {
		t.Errorf("Expected algorithm AES-256-GCM, got %s", encrypted.Algorithm)
	}
	
	if encrypted.Iterations != 1000 {
		t.Errorf("Expected 1000 iterations, got %d", encrypted.Iterations)
	}
	
	if len(encrypted.Data) == 0 {
		t.Error("Encrypted data is empty")
	}
	
	if len(encrypted.Salt) == 0 {
		t.Error("Salt is empty")
	}
	
	if len(encrypted.IV) == 0 {
		t.Error("IV is empty")
	}
}

func TestEncryptor_Decrypt(t *testing.T) {
	encryptor := NewEncryptor(1000)
	
	testData := []byte("test ssh key data")
	passphrase := "test-passphrase"
	
	// Encrypt first
	encrypted, err := encryptor.Encrypt(testData, passphrase)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}
	
	// Decrypt
	decrypted, err := encryptor.Decrypt(encrypted, passphrase)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}
	
	// Verify decrypted data matches original
	if string(decrypted) != string(testData) {
		t.Errorf("Decrypted data doesn't match original. Expected %s, got %s", 
			string(testData), string(decrypted))
	}
}

func TestEncryptor_DecryptWrongPassphrase(t *testing.T) {
	encryptor := NewEncryptor(1000)
	
	testData := []byte("test ssh key data")
	passphrase := "test-passphrase"
	wrongPassphrase := "wrong-passphrase"
	
	// Encrypt with correct passphrase
	encrypted, err := encryptor.Encrypt(testData, passphrase)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}
	
	// Try to decrypt with wrong passphrase
	_, err = encryptor.Decrypt(encrypted, wrongPassphrase)
	if err == nil {
		t.Error("Expected decryption to fail with wrong passphrase")
	}
}

func TestEncryptor_EncryptFiles(t *testing.T) {
	encryptor := NewEncryptor(1000)
	
	files := map[string][]byte{
		"id_rsa":     []byte("private key data"),
		"id_rsa.pub": []byte("public key data"),
		"config":     []byte("ssh config data"),
	}
	passphrase := "test-passphrase"
	
	encrypted, err := encryptor.EncryptFiles(files, passphrase)
	if err != nil {
		t.Fatalf("File encryption failed: %v", err)
	}
	
	if len(encrypted) != len(files) {
		t.Errorf("Expected %d encrypted files, got %d", len(files), len(encrypted))
	}
	
	// Verify each file is encrypted
	for filename := range files {
		if _, exists := encrypted[filename]; !exists {
			t.Errorf("File %s was not encrypted", filename)
		}
	}
}

func TestEncryptor_DecryptFiles(t *testing.T) {
	encryptor := NewEncryptor(1000)
	
	files := map[string][]byte{
		"id_rsa":     []byte("private key data"),
		"id_rsa.pub": []byte("public key data"),
		"config":     []byte("ssh config data"),
	}
	passphrase := "test-passphrase"
	
	// Encrypt files
	encrypted, err := encryptor.EncryptFiles(files, passphrase)
	if err != nil {
		t.Fatalf("File encryption failed: %v", err)
	}
	
	// Decrypt files
	decrypted, err := encryptor.DecryptFiles(encrypted, passphrase)
	if err != nil {
		t.Fatalf("File decryption failed: %v", err)
	}
	
	// Verify all files decrypted correctly
	for filename, originalData := range files {
		decryptedData, exists := decrypted[filename]
		if !exists {
			t.Errorf("File %s was not decrypted", filename)
			continue
		}
		
		if string(decryptedData) != string(originalData) {
			t.Errorf("Decrypted data for %s doesn't match. Expected %s, got %s",
				filename, string(originalData), string(decryptedData))
		}
	}
}

func TestGeneratePassphrase(t *testing.T) {
	// Test default length
	passphrase, err := GeneratePassphrase(32)
	if err != nil {
		t.Fatalf("Failed to generate passphrase: %v", err)
	}
	
	if len(passphrase) != 32 {
		t.Errorf("Expected passphrase length 32, got %d", len(passphrase))
	}
	
	// Test minimum length enforcement
	passphrase2, err := GeneratePassphrase(8)
	if err != nil {
		t.Fatalf("Failed to generate passphrase: %v", err)
	}
	
	if len(passphrase2) != 32 {
		t.Errorf("Expected minimum passphrase length 32, got %d", len(passphrase2))
	}
	
	// Test uniqueness
	passphrase3, err := GeneratePassphrase(32)
	if err != nil {
		t.Fatalf("Failed to generate passphrase: %v", err)
	}
	
	if passphrase == passphrase3 {
		t.Error("Generated passphrases should be unique")
	}
}

func TestEncryptor_VerifyPassphrase(t *testing.T) {
	encryptor := NewEncryptor(1000)
	
	testData := []byte("test data")
	correctPassphrase := "correct-passphrase"
	wrongPassphrase := "wrong-passphrase"
	
	encrypted, err := encryptor.Encrypt(testData, correctPassphrase)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}
	
	// Test correct passphrase
	if !encryptor.VerifyPassphrase(encrypted, correctPassphrase) {
		t.Error("Correct passphrase should verify successfully")
	}
	
	// Test wrong passphrase
	if encryptor.VerifyPassphrase(encrypted, wrongPassphrase) {
		t.Error("Wrong passphrase should fail verification")
	}
}
