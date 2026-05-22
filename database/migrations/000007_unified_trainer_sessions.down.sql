ALTER TABLE class_sessions DROP COLUMN IF EXISTS trainer_id;

-- We don't necessarily want to bring back the broken tables in a down migration 
-- unless specifically requested, but for completeness:
CREATE TYPE slot_status AS ENUM ('available', 'booked', 'canceled');

CREATE TABLE trainer_slots (
    id SERIAL PRIMARY KEY,
    trainer_id INTEGER NOT NULL REFERENCES trainers(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status VARCHAR(50) NOT NULL
);

CREATE TABLE trainer_bookings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    trainer_slot_id INTEGER NOT NULL REFERENCES trainer_slots(id),
    status VARCHAR(50) NOT NULL,
    UNIQUE (trainer_slot_id)
);
