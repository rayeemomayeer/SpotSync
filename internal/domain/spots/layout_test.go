package spots_test

import (
	"testing"

	"github.com/rayeemomayeer/SpotSync/internal/domain/spots"
)

func TestShowcaseLayout(t *testing.T) {
	layout := spots.ShowcaseLayout(1)
	if len(layout) != 24 {
		t.Fatalf("len = %d, want 24", len(layout))
	}
	if layout[0].Label != "EV-01" {
		t.Fatalf("label = %q", layout[0].Label)
	}
}
