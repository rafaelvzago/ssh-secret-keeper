package storage

import (
	"fmt"

	"github.com/rzago/ssh-secret-keeper/internal/config"
	"github.com/rzago/ssh-secret-keeper/internal/interfaces"
)

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) CreateStorage(cfg *config.Config) (interfaces.StorageProvider, error) {
	// Ensure storage provider is set, defaulting to vault for backward compatibility
	if cfg.Storage.Provider == "" {
		cfg.Storage.Provider = "vault"
	}

	switch cfg.Storage.Provider {
	case "vault":
		// Always use the main vault config which has environment variable overrides applied
		// This ensures VAULT_ADDR and VAULT_TOKEN environment variables work correctly
		vaultCfg := &cfg.Vault


		if vaultCfg.Address == "" {
			return nil, fmt.Errorf("vault address not configured - set VAULT_ADDR environment variable or configure vault.address")
		}

		return NewVaultProvider(vaultCfg)

	case "onepassword":
		return nil, fmt.Errorf("1Password provider not implemented yet - coming in future version")

	case "s3":
		return nil, fmt.Errorf("S3 provider not implemented yet - coming in future version")

	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", cfg.Storage.Provider)
	}
}
