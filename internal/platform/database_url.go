package platform

import (
	"net/url"
	"strings"
)

// MigrationDatabaseURL returns the connection string for golang-migrate.
// Prefer DATABASE_MIGRATE_URL when set (Neon direct / non-pooled). Otherwise
// derive from DATABASE_URL by stripping pooler host suffixes unsuitable for DDL.
func MigrationDatabaseURL(databaseURL, migrateOverride string) string {
	if u := strings.TrimSpace(migrateOverride); u != "" {
		return u
	}
	return stripPoolerFromPostgresURL(databaseURL)
}

func stripPoolerFromPostgresURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return strings.ReplaceAll(raw, "-pooler.", ".")
	}

	host := parsed.Hostname()
	if host == "" {
		return raw
	}

	stripped := strings.Replace(host, "-pooler.", ".", 1)
	if stripped == host {
		return raw
	}

	if port := parsed.Port(); port != "" {
		parsed.Host = stripped + ":" + port
	} else {
		parsed.Host = stripped
	}

	return parsed.String()
}
