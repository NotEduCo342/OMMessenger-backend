-- Migration: Create app_versions table for update management
-- Purpose: Track app versions, force updates, and download URLs

CREATE TABLE IF NOT EXISTS app_versions (
    id SERIAL PRIMARY KEY,
    platform VARCHAR(20) NOT NULL,                -- 'android', 'ios', 'web'
    version VARCHAR(20) NOT NULL,                 -- '1.0.0', '1.0.1', etc.
    build_number INTEGER NOT NULL,                -- 1, 2, 3... (CRITICAL: must be unique per platform)
    download_url TEXT NOT NULL,                   -- Full URL to download APK/IPA
    changelog TEXT,                               -- User-facing release notes
    force_update BOOLEAN DEFAULT false,           -- If true, users CANNOT skip
    min_supported_build INTEGER DEFAULT 0,        -- Builds below this are unsupported
    is_active BOOLEAN DEFAULT true,               -- Only ONE active version per platform
    release_date TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT unique_platform_build UNIQUE(platform, build_number),
    CONSTRAINT valid_platform CHECK (platform IN ('android', 'ios', 'web')),
    CONSTRAINT positive_build_number CHECK (build_number > 0),
    CONSTRAINT min_build_valid CHECK (min_supported_build >= 0)
);

-- Indexes for performance
CREATE INDEX idx_app_versions_platform_active ON app_versions(platform, is_active) WHERE is_active = true;
CREATE INDEX idx_app_versions_build_number ON app_versions(platform, build_number DESC);

-- Ensure only ONE active version per platform (critical for consistency)
CREATE UNIQUE INDEX idx_app_versions_one_active 
    ON app_versions(platform) 
    WHERE is_active = true;

-- Trigger to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_app_versions_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER app_versions_update_timestamp
    BEFORE UPDATE ON app_versions
    FOR EACH ROW
    EXECUTE FUNCTION update_app_versions_timestamp();

-- Insert initial Android version (build 1)
INSERT INTO app_versions (
    platform,
    version,
    build_number,
    download_url,
    changelog,
    force_update,
    min_supported_build,
    is_active
) VALUES (
    'android',
    '1.0.0',
    1,
    'https://api-om.wexun.tech/downloads/om-messenger-v1.0.0.apk',
    '• Initial release\n• Offline-first messaging\n• End-to-end encryption\n• Real-time WebSocket communication',
    false,
    1,
    true
);

-- Comment for documentation
COMMENT ON TABLE app_versions IS 'Manages app version information for update notifications';
COMMENT ON COLUMN app_versions.force_update IS 'If true, users cannot dismiss update dialog';
COMMENT ON COLUMN app_versions.min_supported_build IS 'Builds below this number MUST update (can be used with force_update)';
COMMENT ON COLUMN app_versions.is_active IS 'Only one active version per platform - others are historical';
