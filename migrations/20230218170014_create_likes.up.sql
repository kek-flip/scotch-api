CREATE TABLE likes (
    like_id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users,
    liked_user INTEGER REFERENCES users(user_id) CONSTRAINT diff_users CHECK(liked_user != user_id),
    CONSTRAINT must_be_unique_pair UNIQUE(user_id, liked_user)
);