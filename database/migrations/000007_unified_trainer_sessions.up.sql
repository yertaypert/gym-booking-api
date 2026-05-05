ALTER TABLE class_sessions ADD COLUMN trainer_id INTEGER REFERENCES trainers(id);

DROP TABLE IF EXISTS trainer_bookings;
DROP TABLE IF EXISTS trainer_slots;
DROP TYPE IF EXISTS slot_status;
