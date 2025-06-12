DROP TRIGGER IF EXISTS update_support_requests_updated_at ON support_requests;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_support_requests_deleted_at;
DROP INDEX IF EXISTS idx_support_requests_created_at;
DROP INDEX IF EXISTS idx_support_requests_platform;
DROP INDEX IF EXISTS idx_support_requests_status;
DROP INDEX IF EXISTS idx_support_requests_type;
DROP TABLE IF EXISTS support_requests;
