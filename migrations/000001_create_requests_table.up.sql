CREATE TABLE IF NOT EXISTS requests (
    number INTEGER PRIMARY KEY,
    comment TEXT,
    user_id INTEGER,
    username TEXT,
    datetime TEXT
);

CREATE TABLE IF NOT EXISTS NoDellRequests (
    number INTEGER PRIMARY KEY,
    comment TEXT,
    user_id INTEGER,
    username TEXT,
    datetime TEXT
);

CREATE TABLE IF NOT EXISTS traders (
    chat_id INTEGER PRIMARY KEY,
    chat_title TEXT
);

CREATE TABLE IF NOT EXISTS merchants (
    chat_id INTEGER PRIMARY KEY,
    chat_title TEXT
);