CREATE TABLE parking_spots (
    id BIGSERIAL PRIMARY KEY,
    zone_id BIGINT NOT NULL REFERENCES parking_zones (id) ON DELETE CASCADE,
    label VARCHAR(20) NOT NULL,
    pos_x DOUBLE PRECISION NOT NULL,
    pos_y DOUBLE PRECISION NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'available',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT parking_spots_status_check CHECK (status IN ('available', 'unavailable')),
    CONSTRAINT parking_spots_zone_label_unique UNIQUE (zone_id, label)
);

CREATE INDEX idx_parking_spots_zone_id ON parking_spots (zone_id);

ALTER TABLE reservations
    ADD COLUMN spot_id BIGINT REFERENCES parking_spots (id) ON DELETE SET NULL,
    ADD COLUMN demo_expires_at TIMESTAMPTZ;

CREATE UNIQUE INDEX idx_reservations_active_spot_unique
    ON reservations (spot_id)
    WHERE status = 'active' AND spot_id IS NOT NULL;

CREATE INDEX idx_reservations_demo_expires_at
    ON reservations (demo_expires_at)
    WHERE demo_expires_at IS NOT NULL AND status = 'active';
