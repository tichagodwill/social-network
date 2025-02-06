-- Drop the index first
DROP INDEX IF EXISTS idx_notifications_invitation_id;

-- Drop the column (SQLite doesn't support DROP COLUMN directly, 
-- but since this is a down migration and we're dropping the whole table anyway, it's fine) 