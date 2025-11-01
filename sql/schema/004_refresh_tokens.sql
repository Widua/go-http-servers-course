-- +goose Up
CREATE TABLE refresh_tokens(
token text primary key,
created_at TIMESTAMP NOT NULL,
updated_at TIMESTAMP NOT NULL,
expires_at TIMESTAMP NOT NULL,
revoked_at TIMESTAMP,
user_id UUID NOT NULL,
CONSTRAINT fk_userid
	FOREIGN KEY(user_id)
	REFERENCES users(id)
	ON DELETE CASCADE
);

-- +goose Down
DROP TABLE refresh_tokens;
