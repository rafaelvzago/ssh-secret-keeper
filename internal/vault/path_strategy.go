package vault

import (
	"fmt"
	"os"
	"strings"

	"github.com/rzago/ssh-secret-keeper/internal/utils"
)

// StorageStrategy defines different approaches for organizing backups in Vault
type StorageStrategy string

const (
	// StrategyUniversal stores all backups in a shared namespace (default)
	// Path: shared/{namespace}/backups/{backup-name} or shared/backups/{backup-name}
	StrategyUniversal StorageStrategy = "universal"

	// StrategyUser isolates backups by username only (cross-machine compatible)
	// Path: users/{username}/backups/{backup-name}
	StrategyUser StorageStrategy = "user"

	// StrategyMachineUser isolates backups by hostname-username (legacy behavior)
	// Path: users/{hostname-username}/backups/{backup-name}
	StrategyMachineUser StorageStrategy = "machine-user"

	// StrategyCustom uses a user-defined prefix
	// Path: {custom-prefix}/backups/{backup-name}
	StrategyCustom StorageStrategy = "custom"
)

// PathGenerator generates Vault storage paths based on configured strategy
type PathGenerator struct {
	strategy     StorageStrategy
	customPrefix string
	namespace    string
}

// NewPathGenerator creates a new path generator with the specified strategy
func NewPathGenerator(strategy StorageStrategy, customPrefix, namespace string) *PathGenerator {
	return &PathGenerator{
		strategy:     strategy,
		customPrefix: customPrefix,
		namespace:    namespace,
	}
}

// GenerateBasePath creates the base path for backup storage based on the configured strategy
func (p *PathGenerator) GenerateBasePath() (string, error) {
	switch p.strategy {
	case StrategyUniversal:
		return p.generateUniversalPath()
	case StrategyUser:
		return p.generateUserPath()
	case StrategyMachineUser:
		return p.generateMachineUserPath()
	case StrategyCustom:
		return p.generateCustomPath()
	default:
		return "", fmt.Errorf("unknown storage strategy: %s", p.strategy)
	}
}

// generateUniversalPath creates a shared path for universal access
func (p *PathGenerator) generateUniversalPath() (string, error) {
	if p.namespace != "" {
		sanitizedNamespace := utils.SanitizePathComponent(p.namespace)
		return fmt.Sprintf("shared/%s", sanitizedNamespace), nil
	}
	return "shared", nil
}

// generateUserPath creates a user-scoped path (cross-machine compatible)
func (p *PathGenerator) generateUserPath() (string, error) {
	username := p.getCurrentUsername()
	sanitizedUsername := utils.SanitizePathComponent(username)
	return fmt.Sprintf("users/%s", sanitizedUsername), nil
}

// generateMachineUserPath creates the legacy machine-user scoped path
func (p *PathGenerator) generateMachineUserPath() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}

	username := p.getCurrentUsername()

	// Sanitize components for path safety
	sanitizedHostname := utils.SanitizePathComponent(hostname)
	sanitizedUsername := utils.SanitizePathComponent(username)

	return fmt.Sprintf("users/%s-%s", sanitizedHostname, sanitizedUsername), nil
}

// generateCustomPath creates a path with user-defined prefix
func (p *PathGenerator) generateCustomPath() (string, error) {
	if p.customPrefix == "" {
		return "", fmt.Errorf("custom prefix required for custom storage strategy")
	}

	sanitizedPrefix := utils.SanitizePathComponent(p.customPrefix)
	return sanitizedPrefix, nil
}

// getCurrentUsername gets the current username with fallback
func (p *PathGenerator) getCurrentUsername() string {
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME") // Windows fallback
	}
	if username == "" {
		username = "unknown-user"
	}
	return username
}

// GetStrategyDescription returns a human-readable description of the storage strategy
func (p *PathGenerator) GetStrategyDescription() string {
	switch p.strategy {
	case StrategyUniversal:
		if p.namespace != "" {
			return fmt.Sprintf("Universal storage with namespace '%s' (shared across machines and users)", p.namespace)
		}
		return "Universal storage (shared across machines and users)"
	case StrategyUser:
		return "User-scoped storage (isolated by username, shared across machines)"
	case StrategyMachineUser:
		return "Machine-user scoped storage (isolated by hostname and username)"
	case StrategyCustom:
		return fmt.Sprintf("Custom storage with prefix '%s'", p.customPrefix)
	default:
		return "Unknown storage strategy"
	}
}

// ValidateStrategy checks if the strategy configuration is valid
func (p *PathGenerator) ValidateStrategy() error {
	switch p.strategy {
	case StrategyUniversal, StrategyUser, StrategyMachineUser:
		// These strategies don't require additional configuration
		return nil
	case StrategyCustom:
		if p.customPrefix == "" {
			return fmt.Errorf("custom prefix is required for custom storage strategy")
		}
		if strings.Contains(p.customPrefix, "/") {
			return fmt.Errorf("custom prefix cannot contain path separators")
		}
		return nil
	default:
		return fmt.Errorf("unknown storage strategy: %s", p.strategy)
	}
}

// ParseStrategy converts a string to StorageStrategy with validation
func ParseStrategy(strategy string) (StorageStrategy, error) {
	switch strings.ToLower(strings.TrimSpace(strategy)) {
	case "universal", "shared":
		return StrategyUniversal, nil
	case "user":
		return StrategyUser, nil
	case "machine-user", "machine_user", "legacy":
		return StrategyMachineUser, nil
	case "custom":
		return StrategyCustom, nil
	default:
		return "", fmt.Errorf("invalid storage strategy: %s (valid options: universal, user, machine-user, custom)", strategy)
	}
}

// GetAllStrategies returns all available storage strategies with descriptions
func GetAllStrategies() map[StorageStrategy]string {
	return map[StorageStrategy]string{
		StrategyUniversal:   "Universal storage - shared across machines and users (recommended)",
		StrategyUser:        "User-scoped storage - isolated by username, shared across machines",
		StrategyMachineUser: "Machine-user scoped storage - isolated by hostname and username (legacy)",
		StrategyCustom:      "Custom storage - user-defined prefix for advanced scenarios",
	}
}

// MigrationInfo provides information about migrating between storage strategies
type MigrationInfo struct {
	FromStrategy StorageStrategy
	ToStrategy   StorageStrategy
	FromPath     string
	ToPath       string
	Compatible   bool
	Risks        []string
	Benefits     []string
}

// GetMigrationInfo provides detailed information about migrating between strategies
func GetMigrationInfo(from, to StorageStrategy, fromPath, toPath string) *MigrationInfo {
	info := &MigrationInfo{
		FromStrategy: from,
		ToStrategy:   to,
		FromPath:     fromPath,
		ToPath:       toPath,
		Compatible:   true,
		Risks:        []string{},
		Benefits:     []string{},
	}

	// Analyze migration compatibility and implications
	switch {
	case from == StrategyMachineUser && to == StrategyUniversal:
		info.Benefits = []string{
			"Enables cross-machine backup restore",
			"Simplifies backup management",
			"Reduces storage path complexity",
		}
		info.Risks = []string{
			"Backup names must be unique across all machines",
			"Potential conflicts if multiple machines use same backup names",
		}

	case from == StrategyMachineUser && to == StrategyUser:
		info.Benefits = []string{
			"Enables cross-machine backup restore for same user",
			"Maintains user isolation",
		}
		info.Risks = []string{
			"Backup names must be unique across machines for same user",
		}

	case from == StrategyUniversal && to == StrategyUser:
		info.Benefits = []string{
			"Adds user isolation for shared Vault instances",
		}
		info.Risks = []string{
			"Reduces sharing capabilities",
			"May require backup reorganization",
		}

	case from == to:
		info.Benefits = []string{"No migration needed - same strategy"}
		info.Risks = []string{}
	}

	return info
}
