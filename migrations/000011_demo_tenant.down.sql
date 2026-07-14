DROP INDEX IF EXISTS idx_reservations_demo_session;
ALTER TABLE reservations DROP COLUMN IF EXISTS demo_session_id;

DROP INDEX IF EXISTS idx_parking_zones_demo_session;
ALTER TABLE parking_zones DROP COLUMN IF EXISTS demo_session_id;
ALTER TABLE parking_zones DROP COLUMN IF EXISTS is_demo;

ALTER TABLE audit_logs DROP COLUMN IF EXISTS demo_session_id;

ALTER TABLE organizations DROP COLUMN IF EXISTS is_demo;
