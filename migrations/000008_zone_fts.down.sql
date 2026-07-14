DROP TRIGGER IF EXISTS trg_parking_zones_search_vector ON parking_zones;
DROP FUNCTION IF EXISTS parking_zones_search_vector_trigger();
DROP INDEX IF EXISTS idx_parking_zones_search;
ALTER TABLE parking_zones DROP COLUMN IF EXISTS search_vector;
