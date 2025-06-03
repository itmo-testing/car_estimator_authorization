CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY NOT NULL,
    fullName varchar(127) NOT NULL,
    email varchar(32) UNIQUE NOT NULL,
    phone varchar(12) UNIQUE NOT NULL,
    password bytea NOT NULL,
    birthDate date,
    registerDate timestamp with time zone NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS inx_users_id ON users(id);
CREATE INDEX IF NOT EXISTS inx_users_email ON users(email);
