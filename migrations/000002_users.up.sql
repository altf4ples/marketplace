CREATE TABLE store.users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50)                      NOT NULL UNIQUE,
    phone_number VARCHAR(15)                          UNIQUE
);