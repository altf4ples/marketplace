CREATE TABLE store.stores (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(50)                           NOT NULL UNIQUE,
    description VARCHAR(100),
    created_at  TIMESTAMPTZ                           NOT NULL DEFAULT now()
);