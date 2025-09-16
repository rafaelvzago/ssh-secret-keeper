package config

// StorageConfig represents generic storage configuration
type StorageConfig struct {
	Provider    string             `yaml:"provider" mapstructure:"provider"`
	Vault       *VaultConfig       `yaml:"vault,omitempty" mapstructure:"vault"`
	OnePassword *OnePasswordConfig `yaml:"onepassword,omitempty" mapstructure:"onepassword"`
	S3          *S3Config          `yaml:"s3,omitempty" mapstructure:"s3"`
}

// OnePasswordConfig for 1Password Connect API
type OnePasswordConfig struct {
	ServerURL    string `yaml:"server_url" mapstructure:"server_url"`
	Token        string `yaml:"token" mapstructure:"token"`
	VaultID      string `yaml:"vault_id" mapstructure:"vault_id"`
	ItemTemplate string `yaml:"item_template" mapstructure:"item_template"`
}

// S3Config for S3-compatible storage
type S3Config struct {
	Endpoint        string `yaml:"endpoint" mapstructure:"endpoint"`
	Region          string `yaml:"region" mapstructure:"region"`
	Bucket          string `yaml:"bucket" mapstructure:"bucket"`
	AccessKeyID     string `yaml:"access_key_id" mapstructure:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key" mapstructure:"secret_access_key"`
	Prefix          string `yaml:"prefix" mapstructure:"prefix"`
}
