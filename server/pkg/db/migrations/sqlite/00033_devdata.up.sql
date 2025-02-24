-- Enable foreign keys
PRAGMA foreign_keys = ON;

-- Insert into users table
INSERT OR IGNORE INTO users (id, email, username, password, first_name, last_name, date_of_birth, avatar, about_me, is_private, created_at)
VALUES
    (1, 'ali@gmail.com', 'Ali Jasim', '$2a$10$pQTdhZ0jWP1dOViwIan.COMfdwmuhlQlQZa.bVsmfom5V7BL9SpSe', 'Ali', 'Jasim', '2024-01-23 00:00:00+00:00', '', 'me', 0, '2025-02-23 07:37:36'),
    (2, '3@gmail.com', 'latroll', '$2a$10$pQTdhZ0jWP1dOViwIan.COMfdwmuhlQlQZa.bVsmfom5V7BL9SpSe', 'latroll', 'xd', '2025-06-30 00:00:00+00:00', '', 'troll', 0, '2025-02-23 07:39:07'),
    (3, '4@gmail.com', 'anotheruser', '$2a$10$pQTdhZ0jWP1dOViwIan.COMfdwmuhlQlQZa.bVsmfom5V7BL9SpSe', 'another', 'user', '2025-07-01 00:00:00+00:00', '', 'another user', 0, '2025-02-23 07:40:00');

-- Insert into followers table
INSERT OR IGNORE INTO followers (id, follower_id, followed_id, status, created_at)
VALUES
    (1, 2, 1, 'accepted', '2025-02-23 07:50:51'),
    (2, 1, 2, 'accepted', '2025-02-23 07:51:56');

-- Insert into chats table
INSERT OR IGNORE INTO chats (id, type, created_at)
VALUES
    (1, 'direct', '2025-02-23 08:00:00'),
    (2, 'group', '2025-02-23 08:05:00');

-- Insert into chat_messages table
INSERT OR IGNORE INTO chat_messages (id, chat_id, sender_id, content, status, message_type, created_at)
VALUES
    (1, 1, 1, 'Hello!', 'sent', 'text', '2025-02-23 08:01:00'),
    (2, 1, 2, 'Hi there!', 'delivered', 'text', '2025-02-23 08:02:00'),
    (3, 2, 1, 'Group message!', 'read', 'text', '2025-02-23 08:06:00'),
    (4, 2, 3, 'Another group message!', 'sent', 'text', '2025-02-23 08:07:00');

-- Insert into user_chat_status table
INSERT OR IGNORE INTO user_chat_status (id, user_id, chat_id, last_read_message_id)
VALUES
    (1, 1, 1, 1),
    (2, 2, 1, 2),
    (3, 1, 2, 3),
    (4, 3, 2, 4);