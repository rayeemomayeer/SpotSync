-- Postgres FTS for zone name search (additive).
ALTER TABLE parking_zones
    ADD COLUMN IF NOT EXISTS search_vector tsvector;

UPDATE parking_zones
SET search_vector = to_tsvector('english', coalesce(name, ''));

CREATE INDEX IF NOT EXISTS idx_parking_zones_search ON parking_zones USING GIN (search_vector);

CREATE OR REPLACE FUNCTION parking_zones_search_vector_trigger() RETURNS trigger AS $$
BEGIN
    NEW.search_vector := to_tsvector('english', coalesce(NEW.name, ''));
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_parking_zones_search_vector ON parking_zones;
CREATE TRIGGER trg_parking_zones_search_vector
    BEFORE INSERT OR UPDATE OF name ON parking_zones
    FOR EACH ROW EXECUTE FUNCTION parking_zones_search_vector_trigger();
