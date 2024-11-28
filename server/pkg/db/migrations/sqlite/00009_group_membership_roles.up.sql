-- First check if the role column exists
SELECT CASE 
    WHEN COUNT(*) = 0 THEN
        -- If the column doesn't exist, add it
        'ALTER TABLE group_members ADD COLUMN role TEXT DEFAULT "member" NOT NULL;'
    ELSE
        -- If it exists, do nothing
        'SELECT 1;'
END as sql_to_execute
FROM pragma_table_info('group_members')
WHERE name = 'role';

-- Update existing members to have the default role if they don't have one
UPDATE group_members 
SET role = 'member' 
WHERE role IS NULL;

-- Add check constraint if it doesn't exist
SELECT CASE 
    WHEN NOT EXISTS (
        SELECT 1 FROM sqlite_master 
        WHERE type = 'table' 
        AND name = 'group_members' 
        AND sql LIKE '%CHECK (role IN ("member", "moderator", "admin", "creator"))%'
    ) THEN
        'ALTER TABLE group_members ADD CONSTRAINT valid_role CHECK (role IN ("member", "moderator", "admin", "creator"));'
    ELSE
        'SELECT 1;'
END as sql_to_execute;

-- Add invitation system tables
CREATE TABLE group_invitations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    inviter_id INTEGER NOT NULL,
    invitee_id INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' 
        CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (inviter_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (invitee_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(group_id, invitee_id)
);

-- Add indexes
CREATE INDEX idx_group_invitations_group_id ON group_invitations(group_id);
CREATE INDEX idx_group_invitations_invitee_id ON group_invitations(invitee_id);
CREATE INDEX idx_group_invitations_status ON group_invitations(status); 