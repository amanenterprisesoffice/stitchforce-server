-- =========================================
-- 1. ORGANIZATIONS
-- =========================================

CREATE TABLE organizations (
    id TEXT PRIMARY KEY,
    name VARCHAR(200) NOT NULL,

    address_line_1 VARCHAR(255) NOT NULL DEFAULT '',
    address_line_2 VARCHAR(255) NOT NULL DEFAULT '',
    city VARCHAR(100) NOT NULL DEFAULT '',
    state VARCHAR(100) NOT NULL DEFAULT '',
    postal_code VARCHAR(20) NOT NULL DEFAULT '',
    country VARCHAR(100) NOT NULL DEFAULT '',

    photo_key TEXT NOT NULL DEFAULT '',

    license_status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (license_status IN ('active', 'suspended', 'expired')),

    license_expires_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    dashboard_alert_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    dashboard_alert_email VARCHAR(255) NOT NULL DEFAULT '',
    dashboard_alert_schedule JSONB NOT NULL
        DEFAULT '{
            "frequency": "daily",
            "time": "09:00",
            "timezone": "Asia/Kolkata"
        }'::jsonb
    dashboard_alert_last_sent_at TIMESTAMPTZ DEFAULT NULL
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
    photo_key TEXT NOT NULL DEFAULT '',

    role VARCHAR(50) NOT NULL DEFAULT 'staff'
        CHECK (role IN ('org_admin', 'admin', 'staff')),

    email_verified_at TIMESTAMPTZ,
    email_verification_token_hash TEXT,
    email_verification_expires_at TIMESTAMPTZ,

    password_reset_token_hash TEXT NOT NULL DEFAULT '',
    password_reset_token_expires_at TIMESTAMPTZ DEFAULT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,


    max_sessions INTEGER NOT NULL DEFAULT 1
        CHECK (max_sessions >= 1),

    UNIQUE (id, organization_id)
);

CREATE INDEX idx_users_organization_id
    ON users(organization_id);

CREATE UNIQUE INDEX uq_users_organization_email_active
    ON users(organization_id, email)
    WHERE deleted_at IS NULL;


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
    photo_key TEXT NOT NULL DEFAULT '',

    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,

    UNIQUE (id, organization_id),

    -- Required by worker_payments composite foreign key.
    UNIQUE (organization_id, id)
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

    -- Stored in paise.
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

    CONSTRAINT fk_bundles_organization
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_bundles_style
        FOREIGN KEY (style_id, organization_id)
        REFERENCES styles(id, organization_id)
        ON DELETE RESTRICT,

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
CREATE UNIQUE INDEX uq_bundles_organization_style_number
    ON bundles(organization_id, style_id, bundle_number)
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

CREATE INDEX IF NOT EXISTS idx_bundles_completed_archive
ON bundles (organization_id, updated_at)
WHERE status = 'completed'
  AND deleted_at IS NULL;


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
    -- Stored in paise.
    rate_per_piece BIGINT NOT NULL
        CHECK (rate_per_piece >= 0),

    -- Calculated automatically in paise.
    total_amount BIGINT GENERATED ALWAYS AS (
        quantity::BIGINT * rate_per_piece
    ) STORED,

    production_date TIMESTAMPTZ NOT NULL,

    remarks VARCHAR(1000) NOT NULL DEFAULT '',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,

    CONSTRAINT fk_production_entries_organization
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_production_entries_bundle
        FOREIGN KEY (bundle_id, organization_id)
        REFERENCES bundles(id, organization_id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_production_entries_worker
        FOREIGN KEY (worker_id, organization_id)
        REFERENCES garment_workers(id, organization_id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_production_entries_operation
        FOREIGN KEY (style_operation_id, organization_id)
        REFERENCES style_operations(id, organization_id)
        ON DELETE RESTRICT,

    -- Required by worker_payment_allocations composite foreign key.
    UNIQUE (organization_id, id)
);


-- One active production entry per bundle operation.
-- A deleted entry allows the operation to be recorded again.

CREATE UNIQUE INDEX idx_production_entries_bundle_operation
    ON production_entries (
        organization_id,
        bundle_id,
        style_operation_id
    )
    WHERE deleted_at IS NULL;


CREATE INDEX idx_production_entries_organization_id
    ON production_entries(organization_id);

CREATE INDEX idx_production_entries_worker_date
    ON production_entries(organization_id, worker_id, production_date)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_production_entries_bundle_id
    ON production_entries(organization_id, bundle_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_production_entries_operation_id
    ON production_entries(organization_id, style_operation_id)
    WHERE deleted_at IS NULL;


-- =========================================
-- 10. WORKER PAYMENTS
-- =========================================

CREATE TABLE worker_payments (
    id TEXT PRIMARY KEY,

    organization_id TEXT NOT NULL
        REFERENCES organizations(id)
        ON DELETE CASCADE,

    worker_id TEXT NOT NULL,

    -- Total payment amount in paise.
    amount BIGINT NOT NULL
        CHECK (amount > 0),

    payment_date TIMESTAMPTZ NOT NULL,

    payment_method VARCHAR(20) NOT NULL
        CHECK (
            payment_method IN (
                'cash',
                'bank',
                'upi',
                'cheque'
            )
        ),

    reference TEXT NOT NULL DEFAULT '',
    remarks TEXT NOT NULL DEFAULT '',

    status VARCHAR(20) NOT NULL DEFAULT 'completed'
        CHECK (status IN ('completed', 'voided')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_worker_payments_worker
        FOREIGN KEY (organization_id, worker_id)
        REFERENCES garment_workers(organization_id, id)
        ON DELETE RESTRICT,

    -- Required by worker_payment_allocations composite foreign key.
    UNIQUE (organization_id, id)
);


CREATE INDEX idx_worker_payments_worker_date
    ON worker_payments(
        organization_id,
        worker_id,
        payment_date
    );

CREATE INDEX idx_worker_payments_status
    ON worker_payments(organization_id, status);


-- =========================================
-- 11. WORKER PAYMENT ALLOCATIONS
-- =========================================

CREATE TABLE worker_payment_allocations (
    id TEXT PRIMARY KEY,

    organization_id TEXT NOT NULL
        REFERENCES organizations(id)
        ON DELETE CASCADE,

    worker_payment_id TEXT NOT NULL,
    production_entry_id TEXT NOT NULL,

    -- Allocated amount in paise.
    amount BIGINT NOT NULL
        CHECK (amount > 0),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_worker_payment_allocations_payment
        FOREIGN KEY (organization_id, worker_payment_id)
        REFERENCES worker_payments(organization_id, id)
        ON DELETE CASCADE,

    CONSTRAINT fk_worker_payment_allocations_production_entry
        FOREIGN KEY (organization_id, production_entry_id)
        REFERENCES production_entries(organization_id, id)
        ON DELETE RESTRICT,

    UNIQUE (
        organization_id,
        worker_payment_id,
        production_entry_id
    )
);


CREATE INDEX idx_worker_payment_allocations_entry
    ON worker_payment_allocations(
        organization_id,
        production_entry_id
    );

CREATE INDEX idx_worker_payment_allocations_payment
    ON worker_payment_allocations(
        organization_id,
        worker_payment_id
    );



CREATE TABLE licenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    license_key_hash TEXT NOT NULL UNIQUE,
    organization_id UUID NOT NULL REFERENCES organizations(id),

    duration_days INT NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'unused'
        CHECK (status IN ('unused', 'activated', 'revoked')),

    activated_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE organization_invitations (
    id TEXT PRIMARY KEY,

    organization_id TEXT NOT NULL
        REFERENCES organizations(id) ON DELETE CASCADE,

    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL
        CHECK (role IN ('org_admin', 'staff')),

    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,

    accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (organization_id, email)
);

CREATE INDEX idx_organization_invitations_token
    ON organization_invitations(token_hash);



CREATE TABLE user_sessions (
    id TEXT PRIMARY KEY,

    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    device_name VARCHAR(255) NOT NULL DEFAULT '',
    ip_address VARCHAR(45) NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',

    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_user_id
ON user_sessions(user_id);

CREATE INDEX idx_user_sessions_organization_id
ON user_sessions(organization_id);

CREATE INDEX idx_user_sessions_active
ON user_sessions(user_id, created_at DESC)
WHERE revoked_at IS NULL;


CREATE TABLE notifications (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    action_url TEXT NOT NULL DEFAULT '',
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_created
ON notifications(user_id, created_at DESC);

CREATE INDEX idx_notifications_user_unread
ON notifications(user_id, created_at DESC)
WHERE read_at IS NULL;

CREATE TABLE reports (
    id TEXT PRIMARY KEY,

    organization_id TEXT NOT NULL
        REFERENCES organizations(id) ON DELETE CASCADE,

    type VARCHAR(50) NOT NULL,

    year INTEGER NOT NULL,
    month INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),

    file_key TEXT NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,

    record_counts JSONB NOT NULL DEFAULT '{}'::jsonb,

    archived_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    cleanup_completed BOOLEAN NOT NULL DEFAULT FALSE,
    expiration_reminder_sent_at TIMESTAMPTZ,

    CONSTRAINT uq_reports_organization_period
        UNIQUE (organization_id, type, year, month)
);

CREATE INDEX idx_reports_organization_archived
    ON reports(organization_id, archived_at DESC);

CREATE INDEX idx_reports_expiration
    ON reports(expires_at);

-- =========================================
-- END OF SCHEMA
-- =========================================