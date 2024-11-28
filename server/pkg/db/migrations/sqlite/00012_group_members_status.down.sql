-- Since SQLite doesn't support dropping columns, we need to recreate the table
CREATE TABLE group_members_new (
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, user_id),
    CHECK (role IN ('member', 'moderator', 'admin', 'creator'))
);

-- Copy data from old table to new table
INSERT INTO group_members_new (group_id, user_id, role)
SELECT group_id, user_id, role FROM group_members;

-- Drop old table and rename new table
DROP TABLE group_members;
ALTER TABLE group_members_new RENAME TO group_members; 