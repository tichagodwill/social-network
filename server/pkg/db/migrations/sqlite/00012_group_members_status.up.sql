-- Add status column to group_members if it doesn't exist
SELECT CASE 
    WHEN COUNT(*) = 0 THEN
        'ALTER TABLE group_members ADD COLUMN status TEXT NOT NULL DEFAULT "active"'
    ELSE
        'SELECT 1'
END as sql_to_execute
FROM pragma_table_info('group_members')
WHERE name = 'status';

-- Add check constraint for status
SELECT CASE 
    WHEN NOT EXISTS (
        SELECT 1 FROM sqlite_master 
        WHERE type = 'table' 
        AND name = 'group_members' 
        AND sql LIKE '%CHECK (status IN ("active", "inactive", "banned"))%'
    ) THEN
        'ALTER TABLE group_members ADD CONSTRAINT valid_status CHECK (status IN ("active", "inactive", "banned"))'
    ELSE
        'SELECT 1'
END as sql_to_execute; 