CREATE TABLE IF NOT EXISTS contact
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL,
    last_name  TEXT NOT NULL,
    CONSTRAINT contact_first_name_last_name_key UNIQUE (first_name, last_name)
);
