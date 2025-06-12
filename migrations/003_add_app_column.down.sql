-- Remove app column from support_requests table
DROP INDEX IF EXISTS idx_support_requests_app;
ALTER TABLE support_requests DROP COLUMN IF EXISTS app;
