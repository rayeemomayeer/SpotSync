ALTER TABLE organizations
    DROP COLUMN IF EXISTS stripe_customer_id,
    DROP COLUMN IF EXISTS billing_plan;
