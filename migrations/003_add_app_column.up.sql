-- Add app column to support_requests table
ALTER TABLE support_requests ADD COLUMN app VARCHAR(100) NOT NULL DEFAULT 'unknown-app';

-- Create index for the app column for better query performance
CREATE INDEX IF NOT EXISTS idx_support_requests_app ON support_requests(app);
