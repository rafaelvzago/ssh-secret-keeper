package analyzer

import (
	"os"
	"time"
)

// KeyType represents the type of SSH file
type KeyType string

const (
	KeyTypePrivate     KeyType = "private_key"
	KeyTypePublic      KeyType = "public_key"
	KeyTypeConfig      KeyType = "config"
	KeyTypeHosts       KeyType = "known_hosts"
	KeyTypeAuthorized  KeyType = "authorized_keys"
	KeyTypeCertificate KeyType = "certificate"
	KeyTypeUnknown     KeyType = "unknown"
)

// KeyFormat represents the format/encoding of the key
type KeyFormat string

const (
	FormatRSA     KeyFormat = "rsa"
	FormatPEM     KeyFormat = "pem"
	FormatOpenSSH KeyFormat = "openssh"
	FormatConfig  KeyFormat = "ssh_config"
	FormatHosts   KeyFormat = "hosts"
	FormatEd25519 KeyFormat = "ed25519"
	FormatECDSA   KeyFormat = "ecdsa"
	FormatUnknown KeyFormat = "unknown"
)

// KeyPurpose represents the intended use of the key
type KeyPurpose string

const (
	PurposeService  KeyPurpose = "service"
	PurposePersonal KeyPurpose = "personal"
	PurposeWork     KeyPurpose = "work"
	PurposeCloud    KeyPurpose = "cloud"
	PurposeSystem   KeyPurpose = "system"
)

// KeyInfo contains detailed information about an SSH file
type KeyInfo struct {
	Filename    string                 `json:"filename"`
	Type        KeyType                `json:"type"`
	Format      KeyFormat              `json:"format"`
	Service     string                 `json:"service,omitempty"`
	Purpose     KeyPurpose             `json:"purpose"`
	KeyPair     *KeyPairInfo           `json:"key_pair,omitempty"`
	Permissions os.FileMode            `json:"permissions"`
	Size        int64                  `json:"size"`
	ModTime     time.Time              `json:"mod_time"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// KeyPairInfo represents a related key pair
type KeyPairInfo struct {
	BaseName       string   `json:"base_name"`
	PrivateKeyFile string   `json:"private_key_file,omitempty"`
	PublicKeyFile  string   `json:"public_key_file,omitempty"`
	RelatedFiles   []string `json:"related_files,omitempty"`
}

// DetectionResult represents the result of file analysis
type DetectionResult struct {
	Keys         []KeyInfo               `json:"keys"`
	KeyPairs     map[string]*KeyPairInfo `json:"key_pairs"`
	Categories   map[string][]KeyInfo    `json:"categories"`
	SystemFiles  []KeyInfo               `json:"system_files"`
	UnknownFiles []KeyInfo               `json:"unknown_files"`
	Summary      *AnalysisSummary        `json:"summary"`
}

// AnalysisSummary provides a summary of the analysis
type AnalysisSummary struct {
	TotalFiles       int                `json:"total_files"`
	KeyPairCount     int                `json:"key_pair_count"`
	ServiceKeys      int                `json:"service_keys"`
	PersonalKeys     int                `json:"personal_keys"`
	WorkKeys         int                `json:"work_keys"`
	SystemFiles      int                `json:"system_files"`
	UnknownFiles     int                `json:"unknown_files"`
	FormatBreakdown  map[KeyFormat]int  `json:"format_breakdown"`
	PurposeBreakdown map[KeyPurpose]int `json:"purpose_breakdown"`
}

// KeyDetector interface for pluggable key detection
type KeyDetector interface {
	Name() string
	Detect(filename string, content []byte) (*KeyInfo, bool)
	GetRelatedFiles(keyInfo *KeyInfo, allFiles []string) []string
}
