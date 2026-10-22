-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    hashed_password TEXT NOT NULL DEFAULT 'unset',
    email TEXT NOT NULL UNIQUE,
    is_chirpy_red BOOLEAN DEFAULT 'false'    
);

-- +goose Down
DROP TABLE users;


