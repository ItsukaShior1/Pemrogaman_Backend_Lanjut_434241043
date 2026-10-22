CREATE TABLE IF NOT EXISTS users (
    id            SERIAL        PRIMARY KEY,
    username      VARCHAR(64)   NOT NULL,
    email         VARCHAR(255)  NOT NULL,
    password_hash VARCHAR(255)  NOT NULL,
    role          VARCHAR(32)   NOT NULL DEFAULT 'user'
                  CHECK (role IN ('user', 'admin')),
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS users_username_key
    ON users (username);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_key
    ON users (email);

ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
ADD CONSTRAINT users_role_check
CHECK (role IN ('user', 'staff', 'admin'));

ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_role_fkey;

ALTER TABLE users
ADD CONSTRAINT users_role_fkey
FOREIGN KEY (role) REFERENCES roles(name)
ON UPDATE CASCADE;