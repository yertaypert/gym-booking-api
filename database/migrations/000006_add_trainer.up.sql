ALTER TYPE user_role ADD VALUE 'trainer';
CREATE TABLE trainers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    specialization VARCHAR(50),
    extra_fee NUMERIC(10,2) NOT NULL
);

CREATE TABLE trainer_slots (
    id SERIAL PRIMARY KEY,
    trainer_id INTEGER NOT NULL REFERENCES trainers(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT  'available'
);
CREATE TYPE slot_status AS ENUM ('available', 'booked', 'canceled');

CREATE TABLE trainer_bookings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    trainer_slot_id INTEGER NOT NULL REFERENCES trainer_slots(id),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (trainer_slot_id)
);

-- CREATE TYPE booking_status AS ENUM ('active', 'cancelled', 'completed');
