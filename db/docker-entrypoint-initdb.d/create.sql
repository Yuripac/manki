DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    name TEXT NOT NULL,
    email varchar(100) NOT NULL,
    password_encrypted TEXT NOT NULL,
    created_at TIMESTAMP,
    UNIQUE (email)
);

DROP TABLE IF EXISTS cards;
CREATE TABLE cards (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    sentence TEXT NOT NULL,
    meaning TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    repetitions INTEGER DEFAULT 0 NOT NULL,
    efactor REAL DEFAULT 2.5 NOT NULL,
    next_repetition_at DATE,
    created_at TIMESTAMP
);