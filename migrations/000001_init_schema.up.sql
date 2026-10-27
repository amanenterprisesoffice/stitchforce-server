CREATE TABLE organizations (
    id TEXT PRIMARY KEY,
    name VARCHAR(200) NOT NULL,

    license_status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (license_status IN ('active', 'suspended', 'expired')),

    license_expires_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL
        REFERENCES organizations(id) ON DELETE CASCADE,

    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,

    role VARCHAR(50) NOT NULL DEFAULT 'staff'
        CHECK (role IN ('org_admin', 'staff')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (organization_id, email)
);

CREATE INDEX idx_users_organization_id
    ON users(organization_id);

CREATE TABLE garment_workers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    organization_id TEXT NOT NULL
        REFERENCES organizations(id) ON DELETE CASCADE,

    name VARCHAR(200) NOT NULL,
    phone VARCHAR(20),
    gender VARCHAR(10)
        CHECK (gender IN ('male', 'female', 'other')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_garment_workers_organization_id
    ON garment_workers(organization_id);