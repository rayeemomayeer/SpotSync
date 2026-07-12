UPDATE reservations SET organization_id = NULL WHERE organization_id IS NOT NULL;
UPDATE parking_zones SET organization_id = NULL WHERE organization_id IS NOT NULL;

DROP INDEX IF EXISTS idx_outbox_events_dead_lettered;
ALTER TABLE outbox_events
  DROP COLUMN IF EXISTS dead_lettered_at,
  DROP COLUMN IF EXISTS last_error,
  DROP COLUMN IF EXISTS attempts;

DROP INDEX IF EXISTS idx_audit_logs_actor_created;
DROP INDEX IF EXISTS idx_audit_logs_org_created;
DROP TABLE IF EXISTS audit_logs;

DROP INDEX IF EXISTS idx_reservations_organization_id;
ALTER TABLE reservations DROP COLUMN IF EXISTS organization_id;

DROP INDEX IF EXISTS idx_parking_zones_organization_id;
ALTER TABLE parking_zones DROP COLUMN IF EXISTS organization_id;

DROP INDEX IF EXISTS idx_organization_members_user_id;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('driver', 'admin'));
