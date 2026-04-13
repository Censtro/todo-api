CREATE SCHEMA todoapp;

CREATE TABLE todoapp.users (
    id SERIAL PRIMARY KEY,
    version BIGINT NOT NULL DEFAULT 1,
    full_name VARCHAR(100) NOT NULL CHECK (char_length(full_name) BETWEEN 3 and 100),
    phone_number VARCHAR(15) CHECK(
        phone_number ~ '^\+[0-9]+$'
        AND
        char_length(phone_number) BETWEEN 10 and 15
    )
);

CREATE TABLE todoapp.tasks (
    id SERIAL PRIMARY KEY,
    version BIGINT NOT NULL DEFAULT 1,
    title VARCHAR(100) NOT NULL CHECK(char_length(title) BETWEEN 1 and 100),
    description VARCHAR(100) CHECK(char_length(description) BETWEEN 1 and 100),
    completed BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,

    CHECK (
        (completed=FALSE AND completed_at IS NULL)
        OR
        (completed=TRUE AND competed_at IS NOT NULL AND completed_at >= created_at)
    ),
    
    user_id INTEGER NOT NULL REFERENCES todoapp.users(id)
);