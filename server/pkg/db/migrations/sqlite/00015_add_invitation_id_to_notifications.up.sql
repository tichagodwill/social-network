-- Temporarily disable foreign key checks
PRAGMA foreign_keys = OFF;

-- Create a temporary table with the new structure
CREATE TABLE notifications_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    from_user_id INTEGER,
    group_id INTEGER,
    invitation_id INTEGER NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL,
    FOREIGN KEY (invitation_id) REFERENCES group_invitations(id) ON DELETE SET NULL
);

-- Copy existing data with NULL for invitation_id
INSERT INTO notifications_new (id, user_id, type, content, from_user_id, group_id, invitation_id, is_read, created_at)
SELECT 
    id, 
    user_id, 
    type, 
    content, 
    from_user_id, 
    group_id,
    NULL as invitation_id,
    is_read, 
    created_at
FROM notifications;

-- Drop the old table
DROP TABLE notifications;

-- Rename the new table
ALTER TABLE notifications_new RENAME TO notifications;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_invitation_id ON notifications(invitation_id);

-- Re-enable foreign key checks
PRAGMA foreign_keys = ON; 