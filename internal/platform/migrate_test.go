package platform

import (
	"testing"
)

func TestEmbeddedMigrationsPresent(t *testing.T) {
	files, err := EmbeddedMigrationFiles()
	if err != nil {
		t.Fatalf("EmbeddedMigrationFiles() error = %v", err)
	}

	want := map[string]bool{
		"000001_init_schema.up.sql":   false,
		"000001_init_schema.down.sql": false,
	}

	for _, f := range files {
		if _, ok := want[f]; ok {
			want[f] = true
		}
	}

	for name, found := range want {
		if !found {
			t.Errorf("missing embedded migration %q", name)
		}
	}
}
