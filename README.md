Goroku
===
Hypertext dictionary in Go.

# Creating databases
Goroku uses PostgreSQL to keep records of users, captions and entries.

## To create the database;

CREATE DATABASE goroku;

psql goroku
CREATE USER goroku_user WITH PASSWORD 'supersecretpassword987';
GRANT ALL PRIVILEGES ON DATABASE "goroku" to goroku_user;
\q

DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS captions;
DROP TABLE IF EXISTS entries;

CREATE TABLE users (
    user_id INTEGER GENERATED ALWAYS AS IDENTITY,
    username VARCHAR(64) NOT NULL,
    password TEXT NOT NULL,
    email VARCHAR(256) NOT NULL,
    register_date TIMESTAMPTZ NOT NULL,
    last_login TIMESTAMPTZ,
    blocked BOOL,
    PRIMARY KEY(user_id),
    UNIQUE (username, email)
);

CREATE TABLE captions (
    caption_id INTEGER GENERATED ALWAYS AS IDENTITY,
    user_id INTEGER NOT NULL,
    caption VARCHAR(64) NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    updated TIMESTAMPTZ,
    hidden BOOL,
    PRIMARY KEY(caption_id),
    UNIQUE (caption),
    CONSTRAINT fk_user
        FOREIGN KEY(user_id)
            REFERENCES users(user_id)
);

CREATE TABLE entries (
    entry_id INTEGER GENERATED ALWAYS AS IDENTITY,
    caption_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    entry TEXT NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    updated TIMESTAMPTZ,
    hidden BOOL,
    PRIMARY KEY(entry_id),
    CONSTRAINT fk_caption FOREIGN KEY(caption_id) REFERENCES captions(caption_id),
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(user_id)
);