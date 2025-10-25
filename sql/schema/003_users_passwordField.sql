-- +goose Up
ALTER TABLE users add hashed_password text not null;

-- +goose Down
ALTER TABLE users drop column hashed_password;

