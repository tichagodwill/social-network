DROP TABLE IF EXISTS group_event_RSVP;

CREATE TABLE group_event_RSVP (
    event_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    rsvp_status TEXT NOT NULL CHECK(rsvp_status IN ('going', 'not_going')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (event_id) REFERENCES group_events(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (event_id, user_id)
);

CREATE INDEX idx_group_event_rsvp_event_id ON group_event_RSVP(event_id);
CREATE INDEX idx_group_event_rsvp_user_id ON group_event_RSVP(user_id); 