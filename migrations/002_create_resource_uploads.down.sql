-- ==================================================================================
-- DROP RESOURCE UPLOADS MIGRATION
-- ==================================================================================
-- Purpose: Clean rollback of resource_uploads table and related objects
-- Version: 2.0
-- ==================================================================================

-- Drop views first (dependent on table)
DROP VIEW IF EXISTS upload_statistics;
DROP VIEW IF EXISTS active_uploads_monitor;

-- Drop helper functions
DROP FUNCTION IF EXISTS record_part_upload(UUID, INTEGER, TEXT, BIGINT);
DROP FUNCTION IF EXISTS transition_upload_status(UUID, upload_status, TEXT);

-- Drop trigger and trigger function
DROP TRIGGER IF EXISTS update_resource_uploads_update_time ON resource_uploads;
DROP FUNCTION IF EXISTS update_update_time_column();

-- Drop indexes (will be automatically dropped with table, but explicit for clarity)
DROP INDEX IF EXISTS idx_storage_metadata_gin;
DROP INDEX IF EXISTS idx_storage_etags_gin;
DROP INDEX IF EXISTS idx_path_params_gin;
DROP INDEX IF EXISTS idx_stalled_uploads;
DROP INDEX IF EXISTS idx_provider_stats;
DROP INDEX IF EXISTS idx_path_definition;
DROP INDEX IF EXISTS idx_multipart_uploads;
DROP INDEX IF EXISTS idx_pending_cleanup;
DROP INDEX IF EXISTS idx_resource_active_uploads;

-- Drop main table
DROP TABLE IF EXISTS resource_uploads CASCADE;

-- Drop enum types (in reverse dependency order)
DROP TYPE IF EXISTS resource_provider CASCADE;
DROP TYPE IF EXISTS upload_type CASCADE;
DROP TYPE IF EXISTS upload_status CASCADE;