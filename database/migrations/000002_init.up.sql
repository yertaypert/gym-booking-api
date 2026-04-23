CREATE TYPE user_role AS ENUM ('admin', 'user');

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(50) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    full_name VARCHAR(50),
    role user_role DEFAULT 'user',
    balance DECIMAL(10, 2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE gyms (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    address TEXT,
    description TEXT
);

CREATE TABLE classes (
    id SERIAL PRIMARY KEY,
    gym_id INTEGER REFERENCES gyms(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    max_capacity INTEGER NOT NULL
);

CREATE TYPE session_status AS ENUM ('active', 'cancelled', 'completed');

CREATE TABLE class_sessions (
    id SERIAL PRIMARY KEY,
    class_id INTEGER REFERENCES classes(id) ON DELETE CASCADE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    available_slots INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    status session_status DEFAULT 'active'
);

CREATE TYPE booking_status AS ENUM ('pending', 'confirmed', 'attended', 'cancelled');

CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id INTEGER NOT NULL REFERENCES class_sessions(id) ON DELETE CASCADE,
    status booking_status DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_user_session UNIQUE (user_id, session_id)
);

CREATE TYPE transaction_type AS ENUM ('top_up', 'freeze', 'payment', 'refund');

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    booking_id INTEGER REFERENCES bookings(id),
    amount DECIMAL(10, 2) NOT NULL,
    type transaction_type NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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

CREATE TYPE booking_status AS ENUM ('active', 'cancelled', 'completed');