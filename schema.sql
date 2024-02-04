DROP TABLE IF EXISTS TOfficers;
CREATE TABLE TOfficers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT DEFAULT '',
    last_name TEXT DEFAULT '',
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    telegram TEXT DEFAULT '',
    telegram_verified INTEGER DEFAULT 0,
    type TEXT NOT NULL DEFAULT 'user'
);

DROP TABLE IF EXISTS Clients;
CREATE TABLE Clients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    gender TEXT NOT NULL,
    urination INTEGER DEFAULT 300,
    defecation INTEGER DEFAULT 600,
    last_record DATETIME DEFAULT current_timestamp
);

DROP TABLE IF EXISTS Watch;
CREATE TABLE Watch (
    to_id INTEGER,
    client_id INTEGER,
    FOREIGN KEY (to_id) REFERENCES TOfficers (id),
    FOREIGN KEY (client_id) REFERENCES Clients (id),
    UNIQUE (to_id, client_id)
);


