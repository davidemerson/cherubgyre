-- +goose Up
-- +goose StatementBegin
CREATE TABLE profiles (
	id SERIAL PRIMARY KEY,
	private_email VARCHAR(110) NOT NULL UNIQUE,
	profile_image TEXT,
	fullname VARCHAR(255),
	external_id VARCHAR(255),
	created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_profiles_private_email ON profiles(private_email);
CREATE INDEX idx_profiles_external_id ON profiles(external_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_profiles_external_id;
DROP INDEX IF EXISTS idx_profiles_private_email;
DROP TABLE IF EXISTS profiles;
-- +goose StatementEnd
