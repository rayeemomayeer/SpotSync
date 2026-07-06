DROP INDEX IF EXISTS idx_reservations_demo_expires_at;
DROP INDEX IF EXISTS idx_reservations_active_spot_unique;

ALTER TABLE reservations
    DROP COLUMN IF EXISTS demo_expires_at,
    DROP COLUMN IF EXISTS spot_id;

DROP TABLE IF EXISTS parking_spots;
