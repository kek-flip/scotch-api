CREATE TABLE matches (
    match_id SERIAL PRIMARY KEY,
    user_1 INTEGER REFERENCES users(user_id) NOT NULL,
    user_2 INTEGER REFERENCES users(user_id) NOT NULL CONSTRAINT diff_users CHECK(user_1 != user_2),
    UNIQUE (user_1, user_2)
);