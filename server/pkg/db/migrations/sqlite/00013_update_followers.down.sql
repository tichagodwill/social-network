-- Drop indexes
DROP INDEX IF EXISTS idx_followers_status;
DROP INDEX IF EXISTS idx_followers_followed_id;
DROP INDEX IF EXISTS idx_followers_follower_id;

-- Drop table
DROP TABLE IF EXISTS followers; 