PRAGMA foreign_keys = ON;

-- Drop existing tables and indexes if they exist
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS likes;

-- First create the notifications table
CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    from_user_id INTEGER,
    group_id INTEGER,
    is_read BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL
);

-- Then create the indexes after the table exists
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);

-- Create the likes table
CREATE TABLE likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER REFERENCES posts(id),
    comment_id INTEGER REFERENCES comments(id),
    user_id INTEGER REFERENCES users(id),
    is_like BOOLEAN NOT NULL,
    CHECK (post_id IS NOT NULL OR comment_id IS NOT NULL)
);

-- Add indexes for likes table
CREATE INDEX idx_likes_post_id ON likes(post_id);
CREATE INDEX idx_likes_comment_id ON likes(comment_id);
CREATE INDEX idx_likes_user_id ON likes(user_id);
