-- Drop existing tables and indexes
DROP TABLE IF EXISTS group_invitations;
DROP INDEX IF EXISTS idx_group_invitations_status;
DROP INDEX IF EXISTS idx_group_invitations_invitee;
DROP INDEX IF EXISTS idx_group_invitations_group;

-- Create the correct table
CREATE TABLE group_invitations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    inviter_id INTEGER NOT NULL,
    invitee_id INTEGER NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('invitation', 'request')),
    status TEXT NOT NULL CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (inviter_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (invitee_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX idx_group_invitations_group ON group_invitations(group_id);
CREATE INDEX idx_group_invitations_invitee ON group_invitations(invitee_id);
CREATE INDEX idx_group_invitations_status ON group_invitations(status); 