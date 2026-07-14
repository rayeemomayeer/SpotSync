ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS is_demo BOOLEAN NOT NULL DEFAULT false;

UPDATE organizations SET is_demo = true WHERE slug = 'demo-parking';

ALTER TABLE parking_zones
    ADD COLUMN IF NOT EXISTS is_demo BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS demo_session_id VARCHAR(64);

UPDATE parking_zones p
SET is_demo = true
FROM organizations o
WHERE p.organization_id = o.id AND o.is_demo = true;

CREATE INDEX IF NOT EXISTS idx_parking_zones_demo_session ON parking_zones (demo_session_id)
    WHERE demo_session_id IS NOT NULL;

ALTER TABLE reservations
    ADD COLUMN IF NOT EXISTS demo_session_id VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_reservations_demo_session ON reservations (demo_session_id)
    WHERE demo_session_id IS NOT NULL;

ALTER TABLE audit_logs
    ADD COLUMN IF NOT EXISTS demo_session_id VARCHAR(64);
