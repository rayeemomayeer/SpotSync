-- Marketplace tenancy: organizations, org-owned zones, expanded roles, audit, outbox DLQ.

-- Legacy demo role (pre-marketplace) must map before the new CHECK.
UPDATE users SET role = 'admin' WHERE role = 'demo_admin';

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
  CHECK (role IN ('driver', 'admin', 'saas_admin', 'org_admin'));

CREATE TABLE organizations (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT organizations_slug_unique UNIQUE (slug),
    CONSTRAINT organizations_status_check CHECK (status IN ('active', 'suspended'))
);

CREATE TABLE organization_members (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'org_admin',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT organization_members_unique UNIQUE (organization_id, user_id),
    CONSTRAINT organization_members_role_check CHECK (role IN ('org_admin'))
);

CREATE INDEX idx_organization_members_user_id ON organization_members (user_id);

ALTER TABLE parking_zones
  ADD COLUMN IF NOT EXISTS organization_id BIGINT REFERENCES organizations (id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_parking_zones_organization_id ON parking_zones (organization_id);

ALTER TABLE reservations
  ADD COLUMN IF NOT EXISTS organization_id BIGINT REFERENCES organizations (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_reservations_organization_id ON reservations (organization_id);

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    actor_user_id BIGINT REFERENCES users (id) ON DELETE SET NULL,
    organization_id BIGINT REFERENCES organizations (id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id BIGINT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_org_created ON audit_logs (organization_id, created_at DESC);
CREATE INDEX idx_audit_logs_actor_created ON audit_logs (actor_user_id, created_at DESC);

ALTER TABLE outbox_events
  ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_error TEXT,
  ADD COLUMN IF NOT EXISTS dead_lettered_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_outbox_events_dead_lettered
  ON outbox_events (dead_lettered_at)
  WHERE dead_lettered_at IS NOT NULL;

-- Bootstrap default org and attach existing zones for marketplace migration.
INSERT INTO organizations (name, slug, status)
SELECT 'Demo Parking Co', 'demo-parking', 'active'
WHERE NOT EXISTS (SELECT 1 FROM organizations WHERE slug = 'demo-parking');

UPDATE parking_zones z
SET organization_id = o.id
FROM organizations o
WHERE o.slug = 'demo-parking'
  AND z.organization_id IS NULL;

UPDATE reservations r
SET organization_id = z.organization_id
FROM parking_zones z
WHERE r.zone_id = z.id
  AND r.organization_id IS NULL
  AND z.organization_id IS NOT NULL;
