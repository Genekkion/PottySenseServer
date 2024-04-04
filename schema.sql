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

DROP TABLE IF EXISTS Track;
CREATE TABLE Track (
    to_id  INTEGER,
    client_id INTEGER,
    FOREIGN KEY (to_id) REFERENCES TOfficers (id),
    FOREIGN KEY (client_id) REFERENCES Clients (id),
    UNIQUE (to_id, client_id)
);

DROP TABLE IF EXISTS ToiletEntries;
CREATE TABLE ToiletEntries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id INTEGER,
    business_type TEXT NOT NULL,
    duration INTEGER NOT NULL,
    FOREIGN KEY (client_id) REFERENCES Clients (id)
)

DROP TABLE IF EXISTS Toilets;
CREATE TABLE Toilets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    location TEXT NOT NULL
);
