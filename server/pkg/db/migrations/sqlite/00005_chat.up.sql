PRAGMA
foreign_keys = ON;

--- Chat Tables

-- Chats Table
CREATE TABLE chats
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    type       TEXT NOT NULL, -- 'direct' or 'group'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);


-- Chat Messages Table
CREATE TABLE chat_messages
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id      INTEGER NOT NULL REFERENCES chats (id),
    sender_id    INTEGER NOT NULL REFERENCES users (id),
    content      TEXT    NOT NULL,
    status       TEXT     DEFAULT 'sent', -- 'sent', 'delivered', 'read'
    message_type TEXT     DEFAULT 'text', -- 'text', 'file', 'image'
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    CHECK (message_type IN ('text', 'file', 'image')),
    CHECK (status IN ('sent', 'delivered', 'read'))
);

-- User Chat Status Table
CREATE TABLE user_chat_status
(
    id                   INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id              INTEGER NOT NULL REFERENCES users (id),
    chat_id              INTEGER NOT NULL REFERENCES chats (id),
    last_read_message_id INTEGER REFERENCES chat_messages (id),
    UNIQUE (user_id, chat_id)
);

-- Indexes for performance
CREATE INDEX idx_chat_messages_chat_id ON chat_messages (chat_id);
CREATE INDEX idx_chat_messages_sender_id ON chat_messages (sender_id);
CREATE INDEX idx_user_chat_status_user_id ON user_chat_status (user_id);
CREATE INDEX idx_user_chat_status_chat_id ON user_chat_status (chat_id);