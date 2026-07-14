UPDATE organizations SET status = 'active' WHERE status IN ('pending', 'rejected');

ALTER TABLE organizations DROP CONSTRAINT IF EXISTS organizations_status_check;

ALTER TABLE organizations
    ADD CONSTRAINT organizations_status_check
        CHECK (status IN ('active', 'suspended'));
