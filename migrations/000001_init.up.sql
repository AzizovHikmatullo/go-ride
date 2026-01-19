CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL, 
    role TEXT NOT NULL CHECK(role IN ('USER', 'DRIVER')),
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE rides (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    driver_id INTEGER REFERENCES users(id),
    status TEXT NOT NULL CHECK(status IN('SEARCHING', 'IN_PROGRESS', 'COMPLETED', 'CANCELED')),
    start_point JSONB NOT NULL,
    end_point JSONB NOT NULL,
    route JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);