DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    name TEXT NOT NULL
);

DROP TABLE IF EXISTS cards;
CREATE TABLE cards (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    sentence TEXT NOT NULL,
    meaning TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    repetitions INTEGER DEFAULT 0 NOT NULL,
    efactor REAL DEFAULT 2.5 NOT NULL,
    next_repetition_at TEXT,
    FOREIGN KEY(user_id) references users(id)
);