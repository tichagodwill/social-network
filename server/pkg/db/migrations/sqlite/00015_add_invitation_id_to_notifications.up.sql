-- Add invitation_id column to notifications table if it doesn't exist
ALTER TABLE notifications ADD COLUMN invitation_id INTEGER REFERENCES group_invitations(id) ON DELETE CASCADE;

-- Create index for better performance
CREATE INDEX IF NOT EXISTS idx_notifications_invitation_id ON notifications(invitation_id); 