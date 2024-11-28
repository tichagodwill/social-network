DROP INDEX IF EXISTS idx_group_invitations_status;
DROP INDEX IF EXISTS idx_group_invitations_invitee_id;
DROP INDEX IF EXISTS idx_group_invitations_group_id;
DROP TABLE IF EXISTS group_invitations;

-- Since SQLite doesn't support dropping columns, we need to:
-- 1. Create a new table without the role column
-- 2. Copy the data
-- 3. Drop the old table
-- 4. Rename the new table

CREATE TABLE group_members_new (
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, user_id)
);

INSERT INTO group_members_new (group_id, user_id)
SELECT group_id, user_id FROM group_members;

DROP TABLE group_members;

ALTER TABLE group_members_new RENAME TO group_members; 