package analyzer

import (
	"testing"
)

func TestRSAKeyDetector_DetectPrivateKey(t *testing.T) {
	detector := &RSAKeyDetector{}

	privateKeyContent := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1234567890abcdef
-----END RSA PRIVATE KEY-----`)

	keyInfo, detected := detector.Detect("id_rsa", privateKeyContent)

	if !detected {
		t.Error("RSA private key not detected")
	}

	if keyInfo.Type != KeyTypePrivate {
		t.Errorf("Expected KeyTypePrivate, got %s", keyInfo.Type)
	}

	if keyInfo.Format != FormatRSA {
		t.Errorf("Expected FormatRSA, got %s", keyInfo.Format)
	}
}

func TestRSAKeyDetector_DetectPublicKey(t *testing.T) {
	detector := &RSAKeyDetector{}

	publicKeyContent := []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDT1234567890 user@host")

	keyInfo, detected := detector.Detect("id_rsa.pub", publicKeyContent)

	if !detected {
		t.Error("RSA public key not detected")
	}

	if keyInfo.Type != KeyTypePublic {
		t.Errorf("Expected KeyTypePublic, got %s", keyInfo.Type)
	}

	if keyInfo.Format != FormatRSA {
		t.Errorf("Expected FormatRSA, got %s", keyInfo.Format)
	}
}

func TestRSAKeyDetector_DetectByFilename(t *testing.T) {
	detector := &RSAKeyDetector{}

	// Test detection by filename pattern
	keyInfo, detected := detector.Detect("github_rsa", []byte("some content"))

	if !detected {
		t.Error("RSA key not detected by filename")
	}

	if keyInfo.Type != KeyTypePrivate {
		t.Errorf("Expected KeyTypePrivate for filename without .pub, got %s", keyInfo.Type)
	}
}

func TestPEMKeyDetector_DetectPrivateKey(t *testing.T) {
	detector := &PEMKeyDetector{}

	pemContent := []byte(`-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDT1234567890
-----END PRIVATE KEY-----`)

	keyInfo, detected := detector.Detect("key.pem", pemContent)

	if !detected {
		t.Error("PEM private key not detected")
	}

	if keyInfo.Type != KeyTypePrivate {
		t.Errorf("Expected KeyTypePrivate, got %s", keyInfo.Type)
	}

	if keyInfo.Format != FormatPEM {
		t.Errorf("Expected FormatPEM, got %s", keyInfo.Format)
	}
}

func TestPEMKeyDetector_DetectCertificate(t *testing.T) {
	detector := &PEMKeyDetector{}

	certContent := []byte(`-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKoK/heBjcOuMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
-----END CERTIFICATE-----`)

	keyInfo, detected := detector.Detect("cert.pem", certContent)

	if !detected {
		t.Error("PEM certificate not detected")
	}

	if keyInfo.Type != KeyTypeCertificate {
		t.Errorf("Expected KeyTypeCertificate, got %s", keyInfo.Type)
	}
}

func TestOpenSSHKeyDetector_DetectEd25519(t *testing.T) {
	detector := &OpenSSHKeyDetector{}

	ed25519Content := []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI1234567890 user@host")

	keyInfo, detected := detector.Detect("id_ed25519.pub", ed25519Content)

	if !detected {
		t.Error("Ed25519 key not detected")
	}

	if keyInfo.Type != KeyTypePublic {
		t.Errorf("Expected KeyTypePublic, got %s", keyInfo.Type)
	}

	if keyInfo.Format != FormatEd25519 {
		t.Errorf("Expected FormatEd25519, got %s", keyInfo.Format)
	}
}

func TestOpenSSHKeyDetector_DetectECDSA(t *testing.T) {
	detector := &OpenSSHKeyDetector{}

	ecdsaContent := []byte("ssh-ecdsa AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAI1234567890 user@host")

	keyInfo, detected := detector.Detect("id_ecdsa.pub", ecdsaContent)

	if !detected {
		t.Error("ECDSA key not detected")
	}

	if keyInfo.Type != KeyTypePublic {
		t.Errorf("Expected KeyTypePublic, got %s", keyInfo.Type)
	}

	if keyInfo.Format != FormatECDSA {
		t.Errorf("Expected FormatECDSA, got %s", keyInfo.Format)
	}
}

func TestConfigFileDetector_DetectByFilename(t *testing.T) {
	detector := &ConfigFileDetector{}

	keyInfo, detected := detector.Detect("config", []byte("Host github.com\n    HostName github.com"))

	if !detected {
		t.Error("SSH config not detected by filename")
	}

	if keyInfo.Type != KeyTypeConfig {
		t.Errorf("Expected KeyTypeConfig, got %s", keyInfo.Type)
	}

	if keyInfo.Format != FormatConfig {
		t.Errorf("Expected FormatConfig, got %s", keyInfo.Format)
	}
}

func TestConfigFileDetector_DetectByContent(t *testing.T) {
	detector := &ConfigFileDetector{}

	configContent := []byte(`Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/github_rsa`)

	keyInfo, detected := detector.Detect("ssh_config", configContent)

	if !detected {
		t.Error("SSH config not detected by content")
	}

	if keyInfo.Type != KeyTypeConfig {
		t.Errorf("Expected KeyTypeConfig, got %s", keyInfo.Type)
	}
}

func TestKnownHostsDetector_DetectByFilename(t *testing.T) {
	detector := &KnownHostsDetector{}

	keyInfo, detected := detector.Detect("known_hosts", []byte("github.com ssh-rsa AAAAB3NzaC1yc2E..."))

	if !detected {
		t.Error("known_hosts not detected by filename")
	}

	if keyInfo.Type != KeyTypeHosts {
		t.Errorf("Expected KeyTypeHosts, got %s", keyInfo.Type)
	}

	if keyInfo.Format != FormatHosts {
		t.Errorf("Expected FormatHosts, got %s", keyInfo.Format)
	}
}

func TestKnownHostsDetector_DetectByContent(t *testing.T) {
	detector := &KnownHostsDetector{}

	hostsContent := []byte(`github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDT1234567890
gitlab.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDT0987654321`)

	keyInfo, detected := detector.Detect("hosts", hostsContent)

	if !detected {
		t.Error("known_hosts not detected by content")
	}

	if keyInfo.Type != KeyTypeHosts {
		t.Errorf("Expected KeyTypeHosts, got %s", keyInfo.Type)
	}
}

func TestAuthorizedKeysDetector_DetectByFilename(t *testing.T) {
	detector := &AuthorizedKeysDetector{}

	keyInfo, detected := detector.Detect("authorized_keys", []byte("ssh-rsa AAAAB3NzaC1yc2E... user@host"))

	if !detected {
		t.Error("authorized_keys not detected by filename")
	}

	if keyInfo.Type != KeyTypeAuthorized {
		t.Errorf("Expected KeyTypeAuthorized, got %s", keyInfo.Type)
	}

	if keyInfo.Format != FormatOpenSSH {
		t.Errorf("Expected FormatOpenSSH, got %s", keyInfo.Format)
	}
}

func TestAuthorizedKeysDetector_DetectByContent(t *testing.T) {
	detector := &AuthorizedKeysDetector{}

	authContent := []byte(`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDT1234567890 user1@host1
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDT0987654321 user2@host2`)

	keyInfo, detected := detector.Detect("auth_keys", authContent)

	if !detected {
		t.Error("authorized_keys not detected by content")
	}

	if keyInfo.Type != KeyTypeAuthorized {
		t.Errorf("Expected KeyTypeAuthorized, got %s", keyInfo.Type)
	}
}

func TestDetectorNames(t *testing.T) {
	detectors := []KeyDetector{
		&RSAKeyDetector{},
		&PEMKeyDetector{},
		&OpenSSHKeyDetector{},
		&ConfigFileDetector{},
		&KnownHostsDetector{},
		&AuthorizedKeysDetector{},
	}

	expectedNames := []string{
		"rsa",
		"pem",
		"openssh",
		"config",
		"known_hosts",
		"authorized_keys",
	}

	for i, detector := range detectors {
		if detector.Name() != expectedNames[i] {
			t.Errorf("Detector %d: expected name %s, got %s", i, expectedNames[i], detector.Name())
		}
	}
}

func TestRSAKeyDetector_GetRelatedFiles(t *testing.T) {
	detector := &RSAKeyDetector{}

	keyInfo := &KeyInfo{
		Filename: "id_rsa",
		Type:     KeyTypePrivate,
	}

	allFiles := []string{"id_rsa", "id_rsa.pub", "github_rsa", "github_rsa.pub", "config"}

	related := detector.GetRelatedFiles(keyInfo, allFiles)

	if len(related) != 1 {
		t.Errorf("Expected 1 related file, got %d", len(related))
	}

	if len(related) > 0 && related[0] != "id_rsa.pub" {
		t.Errorf("Expected related file id_rsa.pub, got %s", related[0])
	}
}
