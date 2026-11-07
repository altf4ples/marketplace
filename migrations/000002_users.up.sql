CREATE TABLE store.users (
    id       BIGSERIAL PRIMARY KEY,
    username VARCHAR(50)                      NOT NULL UNIQUE,
    phoone_number VARCHAR(11)                          UNIQUE
);