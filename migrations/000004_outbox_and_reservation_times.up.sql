CREATE TABLE outbox_events (
    id BIGSERIAL PRIMARY KEY,
    aggregate_type VARCHAR(64) NOT NULL,
    aggregate_id BIGINT NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX idx_outbox_events_unprocessed
    ON outbox_events (created_at)
    WHERE processed_at IS NULL;

ALTER TABLE reservations
    ADD COLUMN start_time TIMESTAMPTZ,
    ADD COLUMN end_time TIMESTAMPTZ,
    ADD COLUMN idempotency_key VARCHAR(128),
    ADD COLUMN version INTEGER NOT NULL DEFAULT 0;

CREATE UNIQUE INDEX idx_reservations_idempotency_key
    ON reservations (idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX idx_reservations_end_time_active
    ON reservations (end_time)
    WHERE status = 'active' AND end_time IS NOT NULL;
