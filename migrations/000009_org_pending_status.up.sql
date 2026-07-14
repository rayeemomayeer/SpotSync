ALTER TABLE organizations DROP CONSTRAINT IF EXISTS organizations_status_check;

ALTER TABLE organizations
    ADD CONSTRAINT organizations_status_check
        CHECK (status IN ('pending', 'active', 'suspended', 'rejected'));

UPDATE organizations
SET billing_plan = 'starter'
WHERE slug = 'demo-parking' AND billing_plan IS NULL;
