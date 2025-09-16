package storage

import (
	"context"
	"fmt"

	"github.com/rzago/ssh-secret-keeper/internal/config"
	"github.com/rzago/ssh-secret-keeper/internal/vault"
)

// VaultProvider wraps the existing Vault client to implement StorageProvider
type VaultProvider struct {
	client *vault.Client
}

func NewVaultProvider(cfg *config.VaultConfig) (*VaultProvider, error) {
	client, err := vault.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	return &VaultProvider{
		client: client,
	}, nil
}

func (v *VaultProvider) TestConnection(ctx context.Context) error {
	return v.client.TestConnection()
}

func (v *VaultProvider) Close() error {
	v.client.Close()
	return nil
}

func (v *VaultProvider) StoreBackup(ctx context.Context, backupName string, data map[string]interface{}) error {
	return v.client.StoreBackup(backupName, data)
}

func (v *VaultProvider) GetBackup(ctx context.Context, backupName string) (map[string]interface{}, error) {
	return v.client.GetBackup(backupName)
}

func (v *VaultProvider) ListBackups(ctx context.Context) ([]string, error) {
	return v.client.ListBackups()
}

func (v *VaultProvider) DeleteBackup(ctx context.Context, backupName string) error {
	return v.client.DeleteBackup(backupName)
}

func (v *VaultProvider) StoreMetadata(ctx context.Context, metadata map[string]interface{}) error {
	return v.client.StoreMetadata(metadata)
}

func (v *VaultProvider) GetMetadata(ctx context.Context) (map[string]interface{}, error) {
	return v.client.GetMetadata()
}

func (v *VaultProvider) GetProviderType() string {
	return "vault"
}

func (v *VaultProvider) GetBasePath() string {
	return v.client.GetBasePath()
}
