DROP INDEX IF EXISTS idx_reservations_end_time_active;
DROP INDEX IF EXISTS idx_reservations_idempotency_key;

ALTER TABLE reservations
    DROP COLUMN IF EXISTS version,
    DROP COLUMN IF EXISTS idempotency_key,
    DROP COLUMN IF EXISTS end_time,
    DROP COLUMN IF EXISTS start_time;

DROP INDEX IF EXISTS idx_outbox_events_unprocessed;
DROP TABLE IF EXISTS outbox_events;
