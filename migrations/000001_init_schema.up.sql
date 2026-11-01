
-- =========================================
-- 1. ORGANIZATIONS
-- =========================================

CREATE TABLE organizations (
    id TEXT PRIMARY KEY,
    name VARCHAR(200) NOT NULL,

    license_status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (license_status IN ('active', 'suspended', 'expired')),

    license_expires_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- =========================================
-- 2. USERS
-- =========================================

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

    UNIQUE (organization_id, email),
    UNIQUE (id, organization_id)
);

CREATE INDEX idx_users_organization_id
    ON users(organization_id);


-- =========================================
-- 3. GARMENT WORKERS
-- =========================================

CREATE TABLE garment_workers (
    id TEXT PRIMARY KEY,

    organization_id TEXT NOT NULL
        REFERENCES organizations(id) ON DELETE CASCADE,

    name VARCHAR(150) NOT NULL,
    phone VARCHAR(20),
    gender VARCHAR(20),
    photo_key TEXT,

    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,

    UNIQUE (id, organization_id)
);

CREATE INDEX idx_garment_workers_organization_id
    ON garment_workers(organization_id);

CREATE INDEX idx_garment_workers_active
    ON garment_workers(organization_id, status)
    WHERE deleted_at IS NULL;


-- =========================================
-- 4. STYLES
-- =========================================

CREATE TABLE styles (
    id TEXT PRIMARY KEY,

    organization_id TEXT NOT NULL
        REFERENCES organizations(id) ON DELETE CASCADE,

    name VARCHAR(200) NOT NULL,
    style_number VARCHAR(100) NOT NULL,
    description TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (organization_id, style_number),
    UNIQUE (id, organization_id)
);

CREATE INDEX idx_styles_organization_id
    ON styles(organization_id);


-- =========================================
-- 5. STYLE OPERATIONS
-- =========================================

CREATE TABLE style_operations (
    id TEXT PRIMARY KEY,

    organization_id TEXT NOT NULL,
    style_id TEXT NOT NULL,

    name VARCHAR(200) NOT NULL,

    rate_per_piece BIGINT NOT NULL DEFAULT 0
        CHECK (rate_per_piece >= 0),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_style_operations_style
        FOREIGN KEY (style_id, organization_id)
        REFERENCES styles(id, organization_id)
        ON DELETE CASCADE,

    UNIQUE (id, organization_id),
    UNIQUE (id, style_id, organization_id)
);

CREATE INDEX idx_style_operations_organization_id
    ON style_operations(organization_id);

CREATE INDEX idx_style_operations_style_id
    ON style_operations(style_id);


-- =========================================
-- END OF CURRENT SCHEMA
-- =========================================
