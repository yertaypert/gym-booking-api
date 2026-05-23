DROP TABLE IF EXISTS trainer_bookings;
DROP TABLE IF EXISTS trainer_slots;
DROP TABLE IF EXISTS trainers;
DROP TYPE IF EXISTS slot_status;
-- Note: 'trainer' value cannot be removed from user_role enum in Postgres without complex operations.
