
CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    media TEXT, -- Path to image or GIF
    privacy INTEGER CHECK (privacy IN (0, 1, 2)) DEFAULT 0, -- 0: public, 1: private, 2: almost private
    author INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    group_id INTEGER REFERENCES groups(id) -- Optional
);


CREATE TABLE post_PrivateViews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER REFERENCES posts(id),            
    user_id INTEGER REFERENCES users(id)
);

CREATE TABLE likes{
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER REFERENCES users(id),
    post_id INTEGER REFERENCES posts(id),
    comment_id INTEGER REFERENCES comments(id),
    is_liked BOOLEAN DEFAULT FALSE,
}

