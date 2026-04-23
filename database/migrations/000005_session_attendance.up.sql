-- Track when a user physically attended a session.
-- The attended_at column is set by the admin when marking a booking as attended.
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS attended_at TIMESTAMP;

-- Fast lookups: "show me all bookings for this session"
CREATE INDEX IF NOT EXISTS idx_bookings_session_id ON bookings(session_id);

-- Fast lookups: "show me all my bookings"
CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings(user_id);

-- Partial index for active (non-cancelled) bookings — used by the duplicate-check query
CREATE INDEX IF NOT EXISTS idx_bookings_active
    ON bookings(user_id, session_id)
    WHERE status != 'cancelled';
