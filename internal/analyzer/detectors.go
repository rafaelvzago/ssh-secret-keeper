package analyzer

import (
	"bytes"
	"strings"
)

// RSAKeyDetector detects RSA private and public keys
type RSAKeyDetector struct{}

func (d *RSAKeyDetector) Name() string {
	return "rsa"
}

func (d *RSAKeyDetector) Detect(filename string, content []byte) (*KeyInfo, bool) {
	filename = strings.ToLower(filename)

	// Check for RSA private key
	if bytes.Contains(content, []byte("BEGIN RSA PRIVATE KEY")) ||
		bytes.Contains(content, []byte("BEGIN OPENSSH PRIVATE KEY")) {
		return &KeyInfo{
			Type:   KeyTypePrivate,
			Format: FormatRSA,
		}, true
	}

	// Check for RSA public key
	if bytes.HasPrefix(content, []byte("ssh-rsa ")) ||
		strings.HasSuffix(filename, ".pub") {
		return &KeyInfo{
			Type:   KeyTypePublic,
			Format: FormatRSA,
		}, true
	}

	// Check filename patterns
	if strings.Contains(filename, "rsa") && !strings.HasSuffix(filename, ".pub") {
		return &KeyInfo{
			Type:   KeyTypePrivate,
			Format: FormatRSA,
		}, true
	}

	return nil, false
}

func (d *RSAKeyDetector) GetRelatedFiles(keyInfo *KeyInfo, allFiles []string) []string {
	if keyInfo.Type != KeyTypePrivate && keyInfo.Type != KeyTypePublic {
		return nil
	}

	baseName := strings.TrimSuffix(keyInfo.Filename, ".pub")
	var related []string

	for _, file := range allFiles {
		if file == baseName || file == baseName+".pub" {
			if file != keyInfo.Filename {
				related = append(related, file)
			}
		}
	}

	return related
}

// PEMKeyDetector detects PEM format keys and certificates
type PEMKeyDetector struct{}

func (d *PEMKeyDetector) Name() string {
	return "pem"
}

func (d *PEMKeyDetector) Detect(filename string, content []byte) (*KeyInfo, bool) {
	if bytes.Contains(content, []byte("BEGIN PRIVATE KEY")) ||
		bytes.Contains(content, []byte("BEGIN EC PRIVATE KEY")) ||
		bytes.Contains(content, []byte("BEGIN RSA PRIVATE KEY")) ||
		strings.HasSuffix(strings.ToLower(filename), ".pem") {

		keyType := KeyTypePrivate
		if bytes.Contains(content, []byte("BEGIN CERTIFICATE")) {
			keyType = KeyTypeCertificate
		}

		return &KeyInfo{
			Type:   keyType,
			Format: FormatPEM,
		}, true
	}

	return nil, false
}

func (d *PEMKeyDetector) GetRelatedFiles(keyInfo *KeyInfo, allFiles []string) []string {
	return nil // PEM files typically don't have related files
}

// OpenSSHKeyDetector detects OpenSSH format keys
type OpenSSHKeyDetector struct{}

func (d *OpenSSHKeyDetector) Name() string {
	return "openssh"
}

func (d *OpenSSHKeyDetector) Detect(filename string, content []byte) (*KeyInfo, bool) {
	contentStr := string(content)

	// Ed25519 keys
	if bytes.HasPrefix(content, []byte("ssh-ed25519 ")) {
		return &KeyInfo{
			Type:   KeyTypePublic,
			Format: FormatEd25519,
		}, true
	}

	// ECDSA keys
	if bytes.HasPrefix(content, []byte("ssh-ecdsa ")) ||
		strings.Contains(contentStr, "BEGIN EC PRIVATE KEY") {
		keyType := KeyTypePublic
		if strings.Contains(contentStr, "PRIVATE") {
			keyType = KeyTypePrivate
		}
		return &KeyInfo{
			Type:   keyType,
			Format: FormatECDSA,
		}, true
	}

	// Generic OpenSSH private key
	if bytes.Contains(content, []byte("BEGIN OPENSSH PRIVATE KEY")) {
		return &KeyInfo{
			Type:   KeyTypePrivate,
			Format: FormatOpenSSH,
		}, true
	}

	return nil, false
}

func (d *OpenSSHKeyDetector) GetRelatedFiles(keyInfo *KeyInfo, allFiles []string) []string {
	baseName := strings.TrimSuffix(keyInfo.Filename, ".pub")
	var related []string

	for _, file := range allFiles {
		if file == baseName || file == baseName+".pub" {
			if file != keyInfo.Filename {
				related = append(related, file)
			}
		}
	}

	return related
}

// ConfigFileDetector detects SSH config files
type ConfigFileDetector struct{}

func (d *ConfigFileDetector) Name() string {
	return "config"
}

func (d *ConfigFileDetector) Detect(filename string, content []byte) (*KeyInfo, bool) {
	if strings.ToLower(filename) == "config" {
		return &KeyInfo{
			Type:   KeyTypeConfig,
			Format: FormatConfig,
		}, true
	}

	// Check content for SSH config patterns
	contentStr := strings.ToLower(string(content))
	if strings.Contains(contentStr, "host ") ||
		strings.Contains(contentStr, "hostname ") ||
		strings.Contains(contentStr, "identityfile ") {
		return &KeyInfo{
			Type:   KeyTypeConfig,
			Format: FormatConfig,
		}, true
	}

	return nil, false
}

func (d *ConfigFileDetector) GetRelatedFiles(keyInfo *KeyInfo, allFiles []string) []string {
	return nil // Config files don't have related files
}

// KnownHostsDetector detects known_hosts files
type KnownHostsDetector struct{}

func (d *KnownHostsDetector) Name() string {
	return "known_hosts"
}

func (d *KnownHostsDetector) Detect(filename string, content []byte) (*KeyInfo, bool) {
	filename = strings.ToLower(filename)

	if strings.Contains(filename, "known_hosts") {
		return &KeyInfo{
			Type:   KeyTypeHosts,
			Format: FormatHosts,
		}, true
	}

	// Check content pattern (hostname + key)
	lines := bytes.Split(content, []byte("\n"))
	for _, line := range lines {
		if len(line) > 0 && bytes.Contains(line, []byte(" ssh-")) {
			return &KeyInfo{
				Type:   KeyTypeHosts,
				Format: FormatHosts,
			}, true
		}
	}

	return nil, false
}

func (d *KnownHostsDetector) GetRelatedFiles(keyInfo *KeyInfo, allFiles []string) []string {
	var related []string
	for _, file := range allFiles {
		if strings.Contains(strings.ToLower(file), "known_hosts") && file != keyInfo.Filename {
			related = append(related, file)
		}
	}
	return related
}

// AuthorizedKeysDetector detects authorized_keys files
type AuthorizedKeysDetector struct{}

func (d *AuthorizedKeysDetector) Name() string {
	return "authorized_keys"
}

func (d *AuthorizedKeysDetector) Detect(filename string, content []byte) (*KeyInfo, bool) {
	filename = strings.ToLower(filename)

	if strings.Contains(filename, "authorized_keys") {
		return &KeyInfo{
			Type:   KeyTypeAuthorized,
			Format: FormatOpenSSH,
		}, true
	}

	// Check content - should contain public keys
	lines := bytes.Split(content, []byte("\n"))
	sshKeyCount := 0
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("ssh-")) {
			sshKeyCount++
		}
	}

	if sshKeyCount > 0 && len(lines) < 50 { // Heuristic: not too many lines
		return &KeyInfo{
			Type:   KeyTypeAuthorized,
			Format: FormatOpenSSH,
		}, true
	}

	return nil, false
}

func (d *AuthorizedKeysDetector) GetRelatedFiles(keyInfo *KeyInfo, allFiles []string) []string {
	return nil // authorized_keys files don't have related files
}

// UniversalFileDetector detects any file that wasn't caught by other detectors
// This ensures ALL files in .ssh folder are backed up regardless of name
type UniversalFileDetector struct{}

func (d *UniversalFileDetector) Name() string {
	return "universal"
}

func (d *UniversalFileDetector) Detect(filename string, content []byte) (*KeyInfo, bool) {
	// This detector catches everything that other detectors missed
	// It should be the last detector in the list
	return &KeyInfo{
		Type:   KeyTypeUnknown,
		Format: FormatUnknown,
	}, true
}

func (d *UniversalFileDetector) GetRelatedFiles(keyInfo *KeyInfo, allFiles []string) []string {
	return nil // Unknown files don't have predictable related files
}
