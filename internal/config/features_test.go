package config_test

import (
	"testing"

	"github.com/rayeemomayeer/SpotSync/internal/config"
)

func TestFeatureEnabled(t *testing.T) {
	t.Setenv("FEATURE_FLAGS", "stripe_billing, org_search")
	if !config.FeatureEnabled("stripe_billing") {
		t.Fatal("expected stripe_billing enabled")
	}
	if !config.FeatureEnabled("org_search") {
		t.Fatal("expected org_search enabled")
	}
	if config.FeatureEnabled("missing") {
		t.Fatal("missing should be false")
	}
}
