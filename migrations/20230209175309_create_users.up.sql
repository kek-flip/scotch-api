CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    login VARCHAR(25) CONSTRAINT min_login_length CHECK(LENGTH(login) >= 3) NOT NULL UNIQUE,
    encrypted_password VARCHAR(100) CONSTRAINT min_password_length CHECK(LENGTH(encrypted_password) >= 3) NOT NULL,
    name VARCHAR(25) NOT NULL,
    age INTEGER CHECK(age BETWEEN 18 AND 99) NOT NULL,
    gender VARCHAR(6) NOT NULL CHECK(gender = 'male' OR gender = 'female'),
    city VARCHAR(25) NOT NULL,
    phone_number VARCHAR(12) NOT NULL UNIQUE,
    about TEXT
);