CREATE TABLE store.stores (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(50)                           NOT NULL UNIQUE,
    description VARCHAR(100),
    created_at  TIMESTAMPTZ                           NOT NULL DEFAULT now()
);