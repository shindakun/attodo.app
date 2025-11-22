-- Supporter subscription management
-- Store supporter status server-side to prevent self-granting

CREATE TABLE IF NOT EXISTS supporters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    did TEXT NOT NULL UNIQUE,
    handle TEXT NOT NULL,
    email TEXT,
    stripe_customer_id TEXT UNIQUE,
    stripe_subscription_id TEXT UNIQUE,
    plan_type TEXT NOT NULL DEFAULT 'supporter',
    is_active BOOLEAN NOT NULL DEFAULT 1,
    start_date DATETIME NOT NULL,
    end_date DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast lookups by DID (primary query pattern)
CREATE INDEX IF NOT EXISTS idx_supporters_did ON supporters(did);

-- Index for fast lookups by customer ID (webhook lookups)
CREATE INDEX IF NOT EXISTS idx_supporters_customer_id ON supporters(stripe_customer_id);

-- Index for fast lookups by subscription ID (webhook lookups)
CREATE INDEX IF NOT EXISTS idx_supporters_subscription_id ON supporters(stripe_subscription_id);

-- Index for email lookups (customer support)
CREATE INDEX IF NOT EXISTS idx_supporters_email ON supporters(email);

-- Index for active supporters (for analytics)
CREATE INDEX IF NOT EXISTS idx_supporters_active ON supporters(is_active) WHERE is_active = 1;
