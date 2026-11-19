CREATE TABLE store.products (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id     UUID         NOT NULL REFERENCES store.stores(id) ON DELETE CASCADE,
    name         VARCHAR(200)   NOT NULL,
    description  TEXT,
    price        NUMERIC(12, 2) NOT NULL CHECK (price >= 0),
    image_url    TEXT,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_store_id ON store.products(store_id);