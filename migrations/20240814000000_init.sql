-- +goose Up

-- +goose StatementBegin
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(50) NOT NULL,
    is_admin BOOLEAN
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE homes (
    id BIGSERIAL UNIQUE NOT NULL,
    address VARCHAR(120) NOT NULL,
    year INT NOT NULL,
    developer VARCHAR(30) NOT NULL,
    reviewer VARCHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE FLAT_STATUS AS ENUM ('created', 'approved', 'declined', 'on_moderation');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE flats (
    number INT NOT NULL,
    price INT NOT NULL,
    rooms INT NOT NULL,
    home_id INT REFERENCES homes(id) NOT NULL,
    status FLAT_STATUS NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX id_idx ON homes USING hash (id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX status_idx_status ON flats USING hash (status);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX status_idx_home ON flats USING hash (home_id);
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE homes;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE flats;
-- +goose StatementEnd
