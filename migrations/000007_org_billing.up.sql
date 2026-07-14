ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS billing_plan VARCHAR(32),
    ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(255);
