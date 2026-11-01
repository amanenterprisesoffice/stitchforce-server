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
-- 6. SIZES
-- =========================================

CREATE TABLE sizes (
    id TEXT PRIMARY KEY,

    organization_id TEXT NOT NULL
        REFERENCES organizations(id) ON DELETE CASCADE,

    name VARCHAR(50) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (organization_id, name),
    UNIQUE (id, organization_id)
);

CREATE INDEX idx_sizes_organization_id
    ON sizes(organization_id);


-- =========================================
-- 7. BUNDLES
-- =========================================

CREATE TABLE bundles (
    id TEXT PRIMARY KEY,

    organization_id TEXT NOT NULL,

    bundle_number VARCHAR(100) NOT NULL,

    style_id TEXT NOT NULL,
    size_id TEXT NOT NULL,

    quantity INTEGER NOT NULL
        CHECK (quantity > 0),

    -- Flexible factory-specific attributes:
    -- color, shade, fabric_lot, cut_number, etc.
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb,

    status VARCHAR(30) NOT NULL DEFAULT 'created'
        CHECK (
            status IN (
                'created',
                'ready',
                'in_progress',
                'completed',
                'cancelled'
            )
        ),

    remarks VARCHAR(1000) NOT NULL DEFAULT '',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,

    -- Organization ownership
    CONSTRAINT fk_bundles_organization
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON DELETE CASCADE,

    -- Ensure style belongs to the same organization
    CONSTRAINT fk_bundles_style
        FOREIGN KEY (style_id, organization_id)
        REFERENCES styles(id, organization_id)
        ON DELETE RESTRICT,

    -- Ensure size belongs to the same organization
    CONSTRAINT fk_bundles_size
        FOREIGN KEY (size_id, organization_id)
        REFERENCES sizes(id, organization_id)
        ON DELETE RESTRICT,

    UNIQUE (id, organization_id)
);


-- =========================================
-- 8. BUNDLE INDEXES
-- =========================================

-- Bundle number must be unique within an organization.
-- Deleted bundle numbers can be reused.

CREATE UNIQUE INDEX uq_bundles_organization_number
    ON bundles(organization_id, bundle_number)
    WHERE deleted_at IS NULL;


CREATE INDEX idx_bundles_organization_id
    ON bundles(organization_id)
    WHERE deleted_at IS NULL;


CREATE INDEX idx_bundles_style_id
    ON bundles(organization_id, style_id)
    WHERE deleted_at IS NULL;


CREATE INDEX idx_bundles_size_id
    ON bundles(organization_id, size_id)
    WHERE deleted_at IS NULL;


CREATE INDEX idx_bundles_status
    ON bundles(organization_id, status)
    WHERE deleted_at IS NULL;


-- =========================================
-- 9. PRODUCTION ENTRIES
-- =========================================

CREATE TABLE production_entries (
    id TEXT PRIMARY KEY,

    organization_id TEXT NOT NULL,

    bundle_id TEXT NOT NULL,
    worker_id TEXT NOT NULL,

    style_operation_id TEXT NOT NULL,

    quantity INTEGER NOT NULL
        CHECK (quantity > 0),

    -- Snapshot of operation rate at the time
    -- this production entry was recorded.
    rate_per_piece BIGINT NOT NULL
        CHECK (rate_per_piece >= 0),

    -- Calculated as quantity * rate_per_piece.
    total_amount BIGINT GENERATED ALWAYS AS (
        quantity::BIGINT * rate_per_piece
    ) STORED,

    production_date TIMESTAMPTZ NOT NULL,

    remarks VARCHAR(1000) NOT NULL DEFAULT '',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,

    -- Organization ownership
    CONSTRAINT fk_production_entries_organization
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON DELETE CASCADE,

    -- Ensure bundle belongs to the same organization
    CONSTRAINT fk_production_entries_bundle
        FOREIGN KEY (bundle_id, organization_id)
        REFERENCES bundles(id, organization_id)
        ON DELETE RESTRICT,

    -- Ensure worker belongs to the same organization
    CONSTRAINT fk_production_entries_worker
        FOREIGN KEY (worker_id, organization_id)
        REFERENCES garment_workers(id, organization_id)
        ON DELETE RESTRICT,

    -- Ensure operation belongs to the same organization
    CONSTRAINT fk_production_entries_operation
        FOREIGN KEY (style_operation_id, organization_id)
        REFERENCES style_operations(id, organization_id)
        ON DELETE RESTRICT
);

-- =========================================
-- END OF SCHEMA
-- =========================================