CREATE TABLE IF NOT EXISTS achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    icon_path VARCHAR(255),
    banner_path VARCHAR(255),
    category VARCHAR(50),
    points INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_active_achievements ON achievements(is_active);
CREATE INDEX IF NOT EXISTS idx_category ON achievements(category);
CREATE INDEX IF NOT EXISTS idx_created_at ON achievements(created_at DESC);