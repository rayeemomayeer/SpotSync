# SQL migrations

Versioned `golang-migrate` up/down files. Schema source of truth — do not rely on `AutoMigrate` for production.

Apply via `MIGRATE_ON_STARTUP=true` or `make migrate`.
