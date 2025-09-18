package storage

import (
	"context"
	"fmt"

	"github.com/rzago/ssh-secret-keeper/internal/config"
	"github.com/rzago/ssh-secret-keeper/internal/vault"
)

// VaultProvider wraps the Vault storage service to implement StorageProvider
type VaultProvider struct {
	service *vault.StorageService
}

func NewVaultProvider(cfg *config.VaultConfig) (*VaultProvider, error) {
	service, err := vault.NewStorageService(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault storage service: %w", err)
	}

	return &VaultProvider{
		service: service,
	}, nil
}

func (v *VaultProvider) TestConnection(ctx context.Context) error {
	return v.service.TestConnection(ctx)
}

func (v *VaultProvider) Close() error {
	v.service.Close()
	return nil
}

func (v *VaultProvider) StoreBackup(ctx context.Context, backupName string, data map[string]interface{}) error {
	return v.service.StoreBackup(ctx, backupName, data)
}

func (v *VaultProvider) GetBackup(ctx context.Context, backupName string) (map[string]interface{}, error) {
	return v.service.GetBackup(ctx, backupName)
}

func (v *VaultProvider) ListBackups(ctx context.Context) ([]string, error) {
	return v.service.ListBackups(ctx)
}

func (v *VaultProvider) DeleteBackup(ctx context.Context, backupName string) error {
	return v.service.DeleteBackup(ctx, backupName)
}

func (v *VaultProvider) StoreMetadata(ctx context.Context, metadata map[string]interface{}) error {
	return v.service.StoreMetadata(ctx, metadata)
}

func (v *VaultProvider) GetMetadata(ctx context.Context) (map[string]interface{}, error) {
	return v.service.GetMetadata(ctx)
}

func (v *VaultProvider) GetProviderType() string {
	return "vault"
}

func (v *VaultProvider) GetBasePath() string {
	return v.service.GetBasePath()
}
