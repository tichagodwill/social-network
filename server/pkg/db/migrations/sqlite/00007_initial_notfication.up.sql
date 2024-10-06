CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    to_user_id INTEGER REFERENCES users(id),             
    content TEXT NOT NULL,    
    from_user_id INTEGER REFERENCES users(id),        
    read BOOLEAN DEFAULT FALSE,            
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    group_id INTEGER REFERENCES groups(id) -- Optional
);
