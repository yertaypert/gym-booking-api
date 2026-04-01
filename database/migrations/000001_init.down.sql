-- Drop in reverse order of dependencies
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS class_sessions;
DROP TABLE IF EXISTS classes;
DROP TABLE IF EXISTS gyms;
DROP TABLE IF EXISTS users;

-- Drop ENUM types
DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS booking_status;
DROP TYPE IF EXISTS session_status;
DROP TYPE IF EXISTS user_role;