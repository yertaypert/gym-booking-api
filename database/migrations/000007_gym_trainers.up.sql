CREATE TABLE gym_trainers (
    gym_id INTEGER NOT NULL REFERENCES gyms(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (gym_id, user_id)
);
