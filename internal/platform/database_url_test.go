package platform

import "testing"

func TestMigrationDatabaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		primary  string
		override string
		want     string
	}{
		{
			name:     "override wins",
			primary:  "postgres://app@ep-foo-pooler.neon.tech/db?sslmode=require",
			override: "postgres://app@ep-foo.neon.tech/db?sslmode=require",
			want:     "postgres://app@ep-foo.neon.tech/db?sslmode=require",
		},
		{
			name:    "strip neon pooler host",
			primary: "postgres://app:secret@ep-cool-darkness-123456-pooler.us-east-2.aws.neon.tech/neondb?sslmode=require",
			want:    "postgres://app:secret@ep-cool-darkness-123456.us-east-2.aws.neon.tech/neondb?sslmode=require",
		},
		{
			name:    "unchanged direct url",
			primary: "postgres://user:pass@localhost:5432/spotsync?sslmode=disable",
			want:    "postgres://user:pass@localhost:5432/spotsync?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MigrationDatabaseURL(tt.primary, tt.override)
			if got != tt.want {
				t.Fatalf("MigrationDatabaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
