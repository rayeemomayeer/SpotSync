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
	if layout[0].Label != "A-01" {
		t.Fatalf("label = %q", layout[0].Label)
	}
}

func TestAppendGridLayout(t *testing.T) {
	all := spots.GridLayout(1, 8)
	added := spots.AppendGridLayout(1, 5, 3)
	if len(added) != 3 {
		t.Fatalf("len=%d", len(added))
	}
	if added[0].Label != all[5].Label {
		t.Fatalf("label=%q want %q", added[0].Label, all[5].Label)
	}
}
