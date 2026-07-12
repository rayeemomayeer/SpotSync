package config

import (
	"os"
	"strings"
)

// FeatureEnabled reports whether a named feature flag is present in FEATURE_FLAGS.
func FeatureEnabled(name string) bool {
	raw := strings.TrimSpace(os.Getenv("FEATURE_FLAGS"))
	if raw == "" {
		return false
	}
	for _, part := range strings.Split(raw, ",") {
		if strings.EqualFold(strings.TrimSpace(part), name) {
			return true
		}
	}
	return false
}
