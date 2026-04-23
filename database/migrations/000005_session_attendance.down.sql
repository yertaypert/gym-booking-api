DROP INDEX IF EXISTS idx_bookings_active;
DROP INDEX IF EXISTS idx_bookings_user_id;
DROP INDEX IF EXISTS idx_bookings_session_id;
ALTER TABLE bookings DROP COLUMN IF EXISTS attended_at;