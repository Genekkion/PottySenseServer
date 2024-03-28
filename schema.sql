DROP TABLE IF EXISTS TOfficers;
CREATE TABLE TOfficers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    first_name TEXT DEFAULT '',
    last_name TEXT DEFAULT '',
    password TEXT NOT NULL,
    telegram_chat_id TEXT DEFAULT '',
    type TEXT NOT NULL DEFAULT 'user'
);

DROP TABLE IF EXISTS Clients;
CREATE TABLE Clients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    gender TEXT NOT NULL,
    urination INTEGER NOT NULL DEFAULT 300,
    defecation INTEGER NOT NULL DEFAULT 600,
    last_record DATETIME NOT NULL DEFAULT current_timestamp
);

DROP TABLE IF EXISTS Watch;
CREATE TABLE Watch (
    to_id NOT NULL INTEGER,
    client_id NOT NULL INTEGER,
    FOREIGN KEY (to_id) REFERENCES TOfficers (id),
    FOREIGN KEY (client_id) REFERENCES Clients (id),
    UNIQUE (to_id, client_id)
);


