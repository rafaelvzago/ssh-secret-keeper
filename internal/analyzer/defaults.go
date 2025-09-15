package analyzer

import (
	"fmt"
	"strings"
)

// getDefaultServicePatterns returns the default service detection patterns
func getDefaultServicePatterns() map[string][]string {
	return map[string][]string{
		"github":     {"*github*", "*gh_*"},
		"gitlab":     {"*gitlab*", "*gl_*"},
		"bitbucket":  {"*bitbucket*", "*bb_*"},
		"argocd":     {"*argocd*", "*argo*"},
		"quay":       {"*quay*"},
		"gke":        {"*gke*", "*gcp*", "*google*"},
		"aws":        {"*aws*", "*ec2*", "*amazon*"},
		"azure":      {"*azure*", "*az_*"},
		"docker":     {"*docker*", "*registry*"},
		"kubernetes": {"*k8s*", "*kube*"},
		"jenkins":    {"*jenkins*", "*ci_*"},
		"terraform":  {"*terraform*", "*tf_*"},
		"ansible":    {"*ansible*"},
		"vault":      {"*vault*", "*hvac*"},
		"consul":     {"*consul*"},
		"nomad":      {"*nomad*"},
	}
}

// getDefaultPurposeRules returns the default purpose detection rules
func getDefaultPurposeRules() map[string]KeyPurpose {
	return map[string]KeyPurpose{
		// Work-related patterns
		"*work*":      PurposeWork,
		"*corp*":      PurposeWork,
		"*company*":   PurposeWork,
		"*office*":    PurposeWork,
		"*business*":  PurposeWork,

		// Personal patterns
		"*personal*": PurposePersonal,
		"*home*":     PurposePersonal,
		"*private*":  PurposePersonal,
		"id_rsa":     PurposePersonal,
		"id_ecdsa":   PurposePersonal,
		"id_ed25519": PurposePersonal,

		// Cloud patterns
		"*cloud*":   PurposeCloud,
		"*gcp*":     PurposeCloud,
		"*aws*":     PurposeCloud,
		"*azure*":   PurposeCloud,
		"*digital*": PurposeCloud,
		"*linode*":  PurposeCloud,
		"*vultr*":   PurposeCloud,
	}
}

// DetectorRegistry manages key detectors following the Open/Closed Principle
type DetectorRegistry struct {
	detectors map[string]KeyDetector
}

// NewDetectorRegistry creates a new detector registry
func NewDetectorRegistry() *DetectorRegistry {
	return &DetectorRegistry{
		detectors: make(map[string]KeyDetector),
	}
}

// Register adds a detector to the registry
func (r *DetectorRegistry) Register(detector KeyDetector) error {
	if detector == nil {
		return fmt.Errorf("cannot register nil detector")
	}

	name := detector.Name()
	if name == "" {
		return fmt.Errorf("detector name cannot be empty")
	}

	// Normalize name (lowercase, no spaces)
	normalizedName := strings.ToLower(strings.ReplaceAll(name, " ", "_"))

	if _, exists := r.detectors[normalizedName]; exists {
		return fmt.Errorf("detector %s already registered", name)
	}

	r.detectors[normalizedName] = detector
	return nil
}

// Unregister removes a detector from the registry
func (r *DetectorRegistry) Unregister(name string) error {
	normalizedName := strings.ToLower(strings.ReplaceAll(name, " ", "_"))

	if _, exists := r.detectors[normalizedName]; !exists {
		return fmt.Errorf("detector %s not found", name)
	}

	delete(r.detectors, normalizedName)
	return nil
}

// Get returns a detector by name
func (r *DetectorRegistry) Get(name string) (KeyDetector, bool) {
	normalizedName := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	detector, exists := r.detectors[normalizedName]
	return detector, exists
}

// GetAll returns all registered detectors
func (r *DetectorRegistry) GetAll() []KeyDetector {
	detectors := make([]KeyDetector, 0, len(r.detectors))
	for _, detector := range r.detectors {
		detectors = append(detectors, detector)
	}
	return detectors
}

// List returns all registered detector names
func (r *DetectorRegistry) List() []string {
	names := make([]string, 0, len(r.detectors))
	for _, detector := range r.detectors {
		names = append(names, detector.Name())
	}
	return names
}

// Count returns the number of registered detectors
func (r *DetectorRegistry) Count() int {
	return len(r.detectors)
}

// Clear removes all detectors from the registry
func (r *DetectorRegistry) Clear() {
	r.detectors = make(map[string]KeyDetector)
}
