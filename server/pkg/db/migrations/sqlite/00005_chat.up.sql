PRAGMA
foreign_keys = ON;

--- Chat Tables

CREATE TABLE chat_messages
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    sender_id    INTEGER REFERENCES users (id),
    recipient_id INTEGER REFERENCES users (id),
    content      TEXT NOT NULL,
    status       VARCHAR(20) DEFAULT 'sent' CHECK (status IN ('sent', 'delivered', 'read')),
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'file', 'image')),
    file_data    BLOB,
    file_name    TEXT,
    file_type    TEXT,
    created_at   DATETIME    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_chat_status
(
    id                   INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id              INTEGER REFERENCES users (id),
    chat_with_id         INTEGER REFERENCES users (id),
    last_read_message_id INTEGER REFERENCES chat_messages (id),
    UNIQUE (user_id, chat_with_id)
);

CREATE TABLE group_messages
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id   INTEGER REFERENCES groups (id),
    user_id    INTEGER REFERENCES users (id),
    content    TEXT NOT NULL,
    media      TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

