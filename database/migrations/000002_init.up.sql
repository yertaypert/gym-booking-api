CREATE TYPE user_role AS ENUM ('admin', 'user');

CREATE TABLE user (
    id SERIAL PRIMARY KEY,
    email VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(50) NOT NULL,
    full_name VARCHAR(50),
    role user_role DEFAULT 'user',
    balance DECIMAL(10, 2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE gym (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    address TEXT,
    description TEXT
);

CREATE TABLE class (
    id SERIAL PRIMARY KEY,
    gym_id INTEGER REFERENCES gyms(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    max_capacity INTEGER NOT NULL
);

CREATE TYPE session_status AS ENUM ('active', 'cancelled', 'completed');

CREATE TABLE class_session (
    id SERIAL PRIMARY KEY,
    class_id INTEGER REFERENCES classes(id) ON DELETE CASCADE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    available_slots INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    status session_status DEFAULT 'active'
);

CREATE TYPE booking_status AS ENUM ('pending', 'confirmed', 'attended', 'cancelled');

CREATE TABLE booking (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    session_id INTEGER REFERENCES class_sessions(id),
    status booking_status DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE transaction_type AS ENUM ('top_up', 'freeze', 'payment', 'refund');

CREATE TABLE transaction (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    booking_id INTEGER REFERENCES bookings(id),
    amount DECIMAL(10, 2) NOT NULL,
    type transaction_type NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);