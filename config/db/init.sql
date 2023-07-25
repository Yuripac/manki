CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE cards (
    id INTEGER PRIMARY KEY,
    sentence TEXT NOT NULL,
    meaning TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    FOREIGN KEY(user_id) references users(id)
);