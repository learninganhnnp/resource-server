-- ==================================================================================
-- RESOURCE UPLOADS TABLE SCHEMA
-- ==================================================================================
-- Purpose: Unified tracking of all resource uploads (simple and multipart)
-- Design: Single table architecture for operational simplicity and atomic operations
-- Version: 2.0
-- ==================================================================================

-- ==================================================================================
-- ENUM TYPES
-- ==================================================================================

-- Upload status lifecycle states
CREATE TYPE upload_status AS ENUM (
    'initializing',  -- Upload record created, preparing URLs
    'pending',       -- URLs generated, waiting for client upload
    'uploading',     -- Client actively uploading (for progress tracking)
    'processing',    -- Post-upload processing (validation, metadata extraction)
    'completing',    -- Finalizing multipart upload with provider
    'completed',     -- Successfully uploaded and verified
    'failed',        -- Upload failed, can be retried
    'aborted'        -- Upload cancelled by user or system
);

-- Upload type classification
CREATE TYPE upload_type AS ENUM (
    'simple',        -- Direct single file upload
    'multipart'      -- Multi-part upload for large files
);

-- Resource provider types
CREATE TYPE resource_provider AS ENUM (
    'cdn',           -- Primary CDN provider
    's3',            -- Amazon S3
    'r2',            -- Cloudflare R2
    'gcs',           -- Google Cloud Storage
    'azure',         -- Azure Blob Storage
    'local'          -- Local filesystem (dev/test)
);


-- ==================================================================================
-- MAIN TABLE: resource_uploads
-- ==================================================================================
CREATE TABLE resource_uploads (
    -- ================================================================================
    -- PRIMARY IDENTIFICATION
    -- ================================================================================
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- ================================================================================
    -- RESOURCE REFERENCE (Links to business entities)
    -- ================================================================================
    resource_type VARCHAR(50) NOT NULL,        -- 'achievement', 'workout', 'avatar', etc.
    resource_id VARCHAR(255) NOT NULL,         -- Foreign key to business table
    resource_field VARCHAR(50),                -- Specific field: 'icon', 'banner', 'thumbnail'
    resource_value VARCHAR(500) NOT NULL,      -- CDN path context from business entity
    resource_provider resource_provider DEFAULT 'cdn',

    -- ================================================================================
    -- UPLOAD TRACKING
    -- ================================================================================
    upload_type upload_type NOT NULL DEFAULT 'simple',
    upload_status upload_status NOT NULL DEFAULT 'initializing',
    upload_error JSONB,                       -- Human-readable status/error message
    
    -- ================================================================================
    -- PATH RESOLUTION (Integration with resource path system)
    -- ================================================================================
    path_definition VARCHAR(100) NOT NULL,     -- Path template name from resource definitions
    path_parameters JSONB DEFAULT '{}',        -- Parameters used for path resolution

    -- ================================================================================
    -- OBJECT STORAGE DETAILS (Populated after successful upload)
    -- ================================================================================
    storage_provider resource_provider NOT NULL,
    storage_key VARCHAR(500) NOT NULL,         -- Storage provider object key
    storage_size BIGINT,                       -- File size in bytes
    storage_metadata JSONB DEFAULT '{}',        -- Storage provider metadata (content_type, content_encoding, cache_control, etc.)
    
    -- ================================================================================
    -- ETAG TRACKING (Unified for simple and multipart)
    -- For simple: {"etag": "abc123"}
    -- For multipart: {"parts": [{"part": 1, "etag": "abc", "size": 5242880}, ...], "final_etag": "xyz"}
    -- ================================================================================
    storage_etags JSONB DEFAULT '{}',

    -- ================================================================================
    -- MULTIPART UPLOAD FIELDS (NULL for simple uploads)
    -- ================================================================================
    storage_multipart_id VARCHAR(255),         -- Provider's multipart upload ID
    total_parts INTEGER,                       -- Expected total number of parts
    uploaded_parts INTEGER,                    -- Number of parts successfully uploaded

    -- ================================================================================
    -- LIFECYCLE TIMESTAMPS
    -- ================================================================================
    create_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_time TIMESTAMPTZ,                  -- When upload actually started
    completed_time TIMESTAMPTZ,                -- When upload completed successfully
    expires_time TIMESTAMPTZ NOT NULL,         -- URL expiration time
    
    -- ================================================================================
    -- CONSTRAINTS
    -- ================================================================================
    CONSTRAINT check_multipart_fields CHECK (
        (upload_type = 'simple' AND storage_multipart_id IS NULL AND total_parts IS NULL AND uploaded_parts IS NULL) OR
        (upload_type = 'multipart' AND storage_multipart_id IS NOT NULL AND total_parts IS NOT NULL AND uploaded_parts IS NOT NULL)
    ),
    CONSTRAINT check_positive_size CHECK (storage_size IS NULL OR storage_size > 0),
    CONSTRAINT check_positive_parts CHECK (total_parts IS NULL OR total_parts > 0),
    CONSTRAINT check_uploaded_parts CHECK (uploaded_parts IS NULL OR (uploaded_parts >= 0 AND  (total_parts IS NULL OR uploaded_parts <= total_parts)))
);

-- ==================================================================================
-- INDEXES FOR PERFORMANCE
-- ==================================================================================

-- Active uploads by resource (excludes completed for smaller index)
CREATE INDEX idx_resource_active_uploads 
ON resource_uploads(resource_type, resource_id, upload_status)
WHERE upload_status NOT IN ('completed', 'aborted');

-- Pending uploads for cleanup job
CREATE INDEX idx_pending_cleanup 
ON resource_uploads(upload_status, expires_time)
WHERE upload_status IN ('pending', 'uploading', 'initializing');

-- Multipart uploads tracking
CREATE INDEX idx_multipart_uploads 
ON resource_uploads(storage_multipart_id, upload_status)
WHERE storage_multipart_id IS NOT NULL;

-- Path resolution lookup
CREATE INDEX idx_path_definition 
ON resource_uploads(path_definition, resource_value);

-- Provider monitoring
CREATE INDEX idx_provider_stats 
ON resource_uploads(storage_provider, upload_status, create_time);

-- Find stalled uploads
CREATE INDEX idx_stalled_uploads 
ON resource_uploads(upload_status, update_time)
WHERE upload_status IN ('uploading', 'processing', 'completing');

-- JSONB indexes for efficient queries
CREATE INDEX idx_path_params_gin ON resource_uploads USING gin(path_parameters);
CREATE INDEX idx_storage_etags_gin ON resource_uploads USING gin(storage_etags);
CREATE INDEX idx_storage_metadata_gin ON resource_uploads USING gin(storage_metadata);

-- ==================================================================================
-- TRIGGERS
-- ==================================================================================

-- Auto-update update_time timestamp
CREATE OR REPLACE FUNCTION update_update_time_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.update_time = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_resource_uploads_update_time
BEFORE UPDATE ON resource_uploads
FOR EACH ROW
EXECUTE FUNCTION update_update_time_column();

-- ==================================================================================
-- HELPER FUNCTIONS
-- ==================================================================================

-- Function to transition upload status with validation
CREATE OR REPLACE FUNCTION transition_upload_status(
    p_upload_id UUID,
    p_new_status upload_status,
    p_status_message TEXT DEFAULT NULL
)
RETURNS BOOLEAN AS $$
DECLARE
    v_current_status upload_status;
    v_upload_type upload_type;
BEGIN
    -- Get current status and type
    SELECT upload_status, upload_type INTO v_current_status, v_upload_type
    FROM resource_uploads
    WHERE id = p_upload_id
    FOR UPDATE;
    
    IF NOT FOUND THEN
        RETURN FALSE;
    END IF;
    
    -- Validate status transitions
    IF v_current_status = 'completed' OR v_current_status = 'aborted' THEN
        -- Terminal states cannot transition
        RETURN FALSE;
    END IF;
    
    -- Update status
    UPDATE resource_uploads
    SET 
        upload_status = p_new_status,
        upload_error = CASE 
            WHEN p_status_message IS NOT NULL 
            THEN jsonb_build_object('message', p_status_message, 'timestamp', CURRENT_TIMESTAMP)
            ELSE upload_error 
        END,
        started_time = CASE 
            WHEN p_new_status = 'uploading' AND started_time IS NULL 
            THEN CURRENT_TIMESTAMP 
            ELSE started_time 
        END,
        completed_time = CASE 
            WHEN p_new_status = 'completed' 
            THEN CURRENT_TIMESTAMP 
            ELSE completed_time 
        END
    WHERE id = p_upload_id;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Function to record part upload for multipart
CREATE OR REPLACE FUNCTION record_part_upload(
    p_upload_id UUID,
    p_part_number INTEGER,
    p_etag TEXT,
    p_part_size BIGINT
)
RETURNS BOOLEAN AS $$
DECLARE
    v_etag_data JSONB;
    v_parts JSONB;
BEGIN
    -- Get current etag data
    SELECT storage_etags INTO v_etag_data
    FROM resource_uploads
    WHERE id = p_upload_id AND upload_type = 'multipart'
    FOR UPDATE;
    
    IF NOT FOUND THEN
        RETURN FALSE;
    END IF;
    
    -- Initialize parts array if not exists
    IF v_etag_data->>'parts' IS NULL THEN
        v_etag_data = jsonb_set(v_etag_data, '{parts}', '[]'::jsonb);
    END IF;
    
    -- Add or update part
    v_parts = v_etag_data->'parts';
    v_parts = v_parts || jsonb_build_object(
        'part', p_part_number,
        'etag', p_etag,
        'size', p_part_size
    );
    
    -- Update record
    UPDATE resource_uploads
    SET 
        storage_etags = jsonb_set(v_etag_data, '{parts}', v_parts),
        uploaded_parts = uploaded_parts + 1
    WHERE id = p_upload_id;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- ==================================================================================
-- VIEWS FOR MONITORING
-- ==================================================================================

-- Active uploads monitoring view
CREATE OR REPLACE VIEW active_uploads_monitor AS
SELECT 
    id,
    resource_type,
    resource_id,
    upload_type,
    upload_status,
    storage_provider,
    CASE 
        WHEN upload_type = 'multipart' 
        THEN ROUND((uploaded_parts::NUMERIC / NULLIF(total_parts, 0)) * 100, 2)
        ELSE NULL
    END AS completion_percentage,
    create_time,
    expires_time,
    EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - create_time)) AS duration_seconds
FROM resource_uploads
WHERE upload_status NOT IN ('completed', 'aborted', 'failed')
ORDER BY create_time DESC;

-- Upload statistics view
CREATE OR REPLACE VIEW upload_statistics AS
SELECT 
    DATE(create_time) AS upload_date,
    resource_type,
    upload_type,
    storage_provider,
    upload_status,
    COUNT(*) AS upload_count,
    AVG(storage_size) AS avg_size_bytes,
    SUM(storage_size) AS total_size_bytes,
    AVG(EXTRACT(EPOCH FROM (completed_time - create_time))) AS avg_duration_seconds
FROM resource_uploads
WHERE create_time >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY DATE(create_time), resource_type, upload_type, storage_provider, upload_status
ORDER BY upload_date DESC, resource_type;