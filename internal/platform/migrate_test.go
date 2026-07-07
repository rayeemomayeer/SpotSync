package platform

import (
	"testing"
)

func TestEmbeddedMigrationsPresent(t *testing.T) {
	files, err := EmbeddedMigrationFiles()
	if err != nil {
		t.Fatalf("EmbeddedMigrationFiles() error = %v", err)
	}

	want := []string{
		"000001_init_schema.up.sql",
		"000001_init_schema.down.sql",
		"000002_parking_spots_and_demo.up.sql",
		"000002_parking_spots_and_demo.down.sql",
		"000003_version_gap_bridge.up.sql",
		"000003_version_gap_bridge.down.sql",
		"000004_outbox_and_reservation_times.up.sql",
		"000004_outbox_and_reservation_times.down.sql",
	}

	found := make(map[string]bool, len(files))
	for _, f := range files {
		found[f] = true
	}

	for _, name := range want {
		if !found[name] {
			t.Errorf("missing embedded migration %q", name)
		}
	}
}
