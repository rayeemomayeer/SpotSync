CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    reservation_id BIGINT REFERENCES reservations (id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    zone_id BIGINT NOT NULL REFERENCES parking_zones (id) ON DELETE RESTRICT,
    stripe_payment_intent_id VARCHAR(255) NOT NULL,
    amount_cents INT NOT NULL CHECK (amount_cents > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'usd',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT payments_stripe_pi_unique UNIQUE (stripe_payment_intent_id),
    CONSTRAINT payments_status_check CHECK (status IN ('pending', 'succeeded', 'failed', 'refunded'))
);

CREATE INDEX idx_payments_user_id ON payments (user_id);
CREATE INDEX idx_payments_reservation_id ON payments (reservation_id);

CREATE TABLE refunds (
    id BIGSERIAL PRIMARY KEY,
    payment_id BIGINT NOT NULL REFERENCES payments (id) ON DELETE RESTRICT,
    stripe_refund_id VARCHAR(255),
    amount_cents INT NOT NULL CHECK (amount_cents > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT refunds_stripe_refund_unique UNIQUE (stripe_refund_id),
    CONSTRAINT refunds_status_check CHECK (status IN ('pending', 'succeeded', 'failed'))
);

CREATE INDEX idx_refunds_payment_id ON refunds (payment_id);
