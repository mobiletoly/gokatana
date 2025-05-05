CREATE SCHEMA IF NOT EXISTS service2_sample;

CREATE TABLE IF NOT EXISTS service2_sample.orders
(
    id            SERIAL PRIMARY KEY,
    order_number  TEXT           NOT NULL UNIQUE,
    customer_name TEXT           NOT NULL,
    total_amount  NUMERIC(10, 2) NOT NULL
);
