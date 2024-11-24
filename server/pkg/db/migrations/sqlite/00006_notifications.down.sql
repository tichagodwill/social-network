-- Drop indexes first
DROP INDEX IF EXISTS idx_notifications_user_id;
DROP INDEX IF EXISTS idx_notifications_created_at;
DROP INDEX IF EXISTS idx_likes_post_id;
DROP INDEX IF EXISTS idx_likes_comment_id;
DROP INDEX IF EXISTS idx_likes_user_id;

-- Then drop tables
DROP TABLE IF EXISTS likes;
DROP TABLE IF EXISTS notifications;