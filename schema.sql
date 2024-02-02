DROP TABLE IF EXISTS TOfficers;
CREATE TABLE TOfficers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT DEFAULT '',
    last_name TEXT DEFAULT '',
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    telegram TEXT DEFAULT '',
    type TEXT NOT NULL DEFAULT "user"
);
/*
DROP TABLE IF EXISTS Groups;
CREATE TABLE Groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);

DROP TABLE IF EXISTS Subscriptions;
CREATE TABLE Subscriptions (
    to_id INTEGER,
    group_id INTEGER,
    FOREIGN KEY (to_id) REFERENCES TOfficers (id),
    FOREIGN KEY (group_id) REFERENCES Groups (id)
);
*/
DROP TABLE IF EXISTS Clients;
CREATE TABLE Clients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    gender TEXT NOT NULL,
    urination INTEGER DEFAULT 300,
    defecation INTEGER DEFAULT 600,
    last_record DATETIME DEFAULT current_timestamp,

    to_id INTEGER DEFAULT -1,
    FOREIGN KEY (to_id) REFERENCES TOfficers (id)
);

DROP TABLE IF EXISTS Watch;
CREATE TABLE Watch (
    to_id INTEGER,
    client_id INTEGER,
    FOREIGN KEY (to_id) REFERENCES TOfficers (id),
    FOREIGN KEY (client_id) REFERENCES Clients (id)
);


