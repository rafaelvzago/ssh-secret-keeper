package types

import (
	"os"
	"time"
)

// BackupOptions defines options for backup operations
type BackupOptions struct {
	Name        string
	SSHDir      string
	Passphrase  string
	DryRun      bool
	Interactive bool
	Filters     []string
}

// RestoreOptions defines options for restore operations
type RestoreOptions struct {
	BackupName  string
	TargetDir   string
	Passphrase  string
	DryRun      bool
	Overwrite   bool
	Interactive bool
	FileFilter  []string
	TypeFilter  []string
}

// BackupResult contains the result of a backup operation
type BackupResult struct {
	Name        string
	FilesBackup int
	TotalSize   int64
	Duration    time.Duration
	StoragePath string
	Checksum    string
}

// RestoreResult contains the result of a restore operation
type RestoreResult struct {
	BackupName    string
	FilesRestored int
	TargetDir     string
	Duration      time.Duration
	SkippedFiles  []string
}

// BackupInfo provides metadata about a backup
type BackupInfo struct {
	Name        string
	Timestamp   time.Time
	FileCount   int
	TotalSize   int64
	Hostname    string
	Username    string
	StoragePath string
}

// FilePermissions represents file permission information
type FilePermissions struct {
	Mode  os.FileMode
	Owner string
	Group string
	Octal string
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
	Code    string
}

func (e ValidationError) Error() string {
	return e.Message
}
