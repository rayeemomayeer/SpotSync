CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'driver',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_role_check CHECK (role IN ('driver', 'admin'))
);

CREATE TABLE parking_zones (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    total_capacity INTEGER NOT NULL,
    price_per_hour NUMERIC(10, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT parking_zones_type_check CHECK (type IN ('general', 'ev_charging', 'covered')),
    CONSTRAINT parking_zones_total_capacity_check CHECK (total_capacity > 0),
    CONSTRAINT parking_zones_price_per_hour_check CHECK (price_per_hour > 0)
);

CREATE TABLE reservations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    zone_id BIGINT NOT NULL REFERENCES parking_zones (id) ON DELETE RESTRICT,
    license_plate VARCHAR(15) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT reservations_status_check CHECK (status IN ('active', 'completed', 'cancelled'))
);

CREATE INDEX idx_reservations_zone_id_status ON reservations (zone_id, status);

CREATE INDEX idx_reservations_user_id ON reservations (user_id);
