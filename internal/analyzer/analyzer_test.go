package analyzer

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAnalyzer_New(t *testing.T) {
	analyzer := New()
	
	if analyzer == nil {
		t.Fatal("New() returned nil")
	}
	
	if len(analyzer.detectors) == 0 {
		t.Error("No detectors registered")
	}
	
	if len(analyzer.servicePatterns) == 0 {
		t.Error("No service patterns configured")
	}
}

func TestAnalyzer_AnalyzeDirectory_NotFound(t *testing.T) {
	analyzer := New()
	
	_, err := analyzer.AnalyzeDirectory("/nonexistent/directory")
	if err == nil {
		t.Error("Expected error for nonexistent directory")
	}
}

func TestAnalyzer_AnalyzeDirectory_Empty(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "ssh-test-empty")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	analyzer := New()
	result, err := analyzer.AnalyzeDirectory(tmpDir)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("Result is nil")
	}
	
	if len(result.Keys) != 0 {
		t.Errorf("Expected 0 keys in empty directory, got %d", len(result.Keys))
	}
}

func TestAnalyzer_AnalyzeDirectory_WithSSHFiles(t *testing.T) {
	// Create temporary SSH directory
	tmpDir, err := os.MkdirTemp("", "ssh-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create test SSH files
	testFiles := map[string]string{
		"id_rsa": `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEA1234567890
-----END OPENSSH PRIVATE KEY-----`,
		"id_rsa.pub": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDT1234567890 user@host",
		"service1_rsa": `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1234567890
-----END RSA PRIVATE KEY-----`,
		"service1_rsa.pub": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDT1234567890 user@host",
		"config": `Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/service1_rsa`,
		"known_hosts": "github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDT1234567890",
	}
	
	for filename, content := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}
	
	analyzer := New()
	result, err := analyzer.AnalyzeDirectory(tmpDir)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}
	
	if len(result.Keys) == 0 {
		t.Error("No keys detected")
	}
	
	// Check key pairs
	if len(result.KeyPairs) != 2 { // id_rsa and service1_rsa pairs
		t.Errorf("Expected 2 key pairs, got %d", len(result.KeyPairs))
	}
	
	// Check categories
	if len(result.Categories) == 0 {
		t.Error("No categories found")
	}
	
	// Check system files
	foundConfig := false
	foundKnownHosts := false
	for _, file := range result.SystemFiles {
		if file.Filename == "config" {
			foundConfig = true
		}
		if file.Filename == "known_hosts" {
			foundKnownHosts = true
		}
	}
	
	if !foundConfig {
		t.Error("SSH config file not detected as system file")
	}
	
	if !foundKnownHosts {
		t.Error("known_hosts file not detected as system file")
	}
}

func TestAnalyzer_GetBaseName(t *testing.T) {
	analyzer := New()
	
	tests := []struct {
		filename string
		expected string
	}{
		{"id_rsa", "id_rsa"},
		{"id_rsa.pub", "id_rsa"},
		{"service1_rsa", "service1_rsa"},
		{"service1_rsa.pub", "service1_rsa"},
		{"test.pem", "test"},
		{"key.rsa", "key"},
		{"simple", "simple"},
	}
	
	for _, test := range tests {
		result := analyzer.getBaseName(test.filename)
		if result != test.expected {
			t.Errorf("getBaseName(%s) = %s, expected %s", test.filename, result, test.expected)
		}
	}
}

func TestDetectionResult_Summary(t *testing.T) {
	// Create a mock detection result
	result := &DetectionResult{
		Keys: []KeyInfo{
			{
				Filename: "id_rsa",
				Type:     KeyTypePrivate,
				Format:   FormatRSA,
				Purpose:  PurposePersonal,
			},
			{
				Filename: "id_rsa.pub",
				Type:     KeyTypePublic,
				Format:   FormatRSA,
				Purpose:  PurposePersonal,
			},
			{
				Filename: "service1_rsa",
				Type:     KeyTypePrivate,
				Format:   FormatRSA,
				Purpose:  PurposeService,
				Service:  "github",
			},
		},
		KeyPairs: map[string]*KeyPairInfo{
			"id_rsa": {
				BaseName:       "id_rsa",
				PrivateKeyFile: "id_rsa",
				PublicKeyFile:  "id_rsa.pub",
			},
		},
	}
	
	// Generate summary
	analyzer := New()
	summary := analyzer.generateSummary(result.Keys, result.KeyPairs)
	
	if summary.TotalFiles != 3 {
		t.Errorf("Expected 3 total files, got %d", summary.TotalFiles)
	}
	
	if summary.KeyPairCount != 1 {
		t.Errorf("Expected 1 key pair, got %d", summary.KeyPairCount)
	}
	
	if summary.PersonalKeys != 2 {
		t.Errorf("Expected 2 personal keys, got %d", summary.PersonalKeys)
	}
	
	if summary.ServiceKeys != 1 {
		t.Errorf("Expected 1 service key, got %d", summary.ServiceKeys)
	}
}

func TestKeyInfo_Fields(t *testing.T) {
	now := time.Now()
	
	keyInfo := KeyInfo{
		Filename:    "test_rsa",
		Type:        KeyTypePrivate,
		Format:      FormatRSA,
		Service:     "github",
		Purpose:     PurposeService,
		Permissions: 0600,
		Size:        1024,
		ModTime:     now,
		Metadata:    map[string]interface{}{"test": "value"},
	}
	
	// Test all fields are set correctly
	if keyInfo.Filename != "test_rsa" {
		t.Errorf("Filename not set correctly")
	}
	
	if keyInfo.Type != KeyTypePrivate {
		t.Errorf("Type not set correctly")
	}
	
	if keyInfo.Format != FormatRSA {
		t.Errorf("Format not set correctly")
	}
	
	if keyInfo.Service != "github" {
		t.Errorf("Service not set correctly")
	}
	
	if keyInfo.Purpose != PurposeService {
		t.Errorf("Purpose not set correctly")
	}
	
	if keyInfo.Permissions != 0600 {
		t.Errorf("Permissions not set correctly")
	}
	
	if keyInfo.Size != 1024 {
		t.Errorf("Size not set correctly")
	}
	
	if !keyInfo.ModTime.Equal(now) {
		t.Errorf("ModTime not set correctly")
	}
	
	if keyInfo.Metadata["test"] != "value" {
		t.Errorf("Metadata not set correctly")
	}
}
