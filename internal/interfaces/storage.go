package interfaces

import (
	"context"
)

// StorageProvider defines the interface for secret storage backends
type StorageProvider interface {
	// Connection management
	TestConnection(ctx context.Context) error
	Close() error

	// Backup operations
	StoreBackup(ctx context.Context, backupName string, data map[string]interface{}) error
	GetBackup(ctx context.Context, backupName string) (map[string]interface{}, error)
	ListBackups(ctx context.Context) ([]string, error)
	DeleteBackup(ctx context.Context, backupName string) error

	// Metadata operations
	StoreMetadata(ctx context.Context, metadata map[string]interface{}) error
	GetMetadata(ctx context.Context) (map[string]interface{}, error)

	// Provider info
	GetProviderType() string
	GetBasePath() string
}

// StorageFactory creates storage providers based on configuration
type StorageFactory interface {
	CreateStorage(cfg interface{}) (StorageProvider, error)
}
