CREATE TABLE trainers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) UNIQUE,
    specialization VARCHAR(50),
    extra_fee NUMERIC(10,2) NOT NULL
);

CREATE TABLE gym_trainers (
    gym_id INTEGER REFERENCES gyms(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (gym_id, user_id)
);

CREATE TABLE trainer_slots (
    id SERIAL PRIMARY KEY,
    trainer_id INTEGER NOT NULL REFERENCES trainers(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status VARCHAR(50) NOT NULL
);

CREATE TYPE slot_status AS ENUM ('available', 'booked', 'canceled');

CREATE TABLE trainer_bookings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    trainer_slot_id INTEGER NOT NULL REFERENCES trainer_slots(id),
    status VARCHAR(50) NOT NULL,
    UNIQUE (trainer_slot_id)
);
