CREATE TABLE IF NOT EXISTS orders
(
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    order_number  TEXT    NOT NULL UNIQUE,
    customer_name TEXT    NOT NULL,
    total_amount  NUMERIC NOT NULL
);