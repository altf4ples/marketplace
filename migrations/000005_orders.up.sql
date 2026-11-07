CREATE TABLE store.orders (
    id           BIGSERIAL      PRIMARY KEY,
    user_id      BIGINT         NOT NULL REFERENCES store.users(id),
    status       VARCHAR(20)    NOT NULL CHECK (status IN (
                     'created', 'assembling', 'shipped', 'delivered', 'cancelled'
                 )),
    total_price  NUMERIC(12, 2) NOT NULL CHECK (total_price >= 0),
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user_id ON store.orders(user_id);
CREATE INDEX idx_orders_status  ON store.orders(status);

CREATE TABLE store.order_items (
    id          BIGSERIAL      PRIMARY KEY,
    order_id    BIGINT         NOT NULL REFERENCES store.orders(id)   ON DELETE CASCADE,
    product_id  BIGINT         NOT NULL REFERENCES store.products(id),
    store_id    BIGINT         NOT NULL REFERENCES store.stores(id),
    quantity    INT            NOT NULL CHECK (quantity > 0),
    price       NUMERIC(12, 2) NOT NULL CHECK (price >= 0)
);

CREATE INDEX idx_order_items_order_id ON store.order_items(order_id);
CREATE INDEX idx_order_items_store_id ON store.order_items(store_id);

CREATE TABLE store.order_state_history (
    id          BIGSERIAL    PRIMARY KEY,
    order_id    BIGINT       NOT NULL REFERENCES store.orders(id) ON DELETE CASCADE,
    from_status VARCHAR(20),                
    to_status   VARCHAR(20)  NOT NULL,
    event       VARCHAR(20)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_order_state_history_order_id ON store.order_state_history(order_id);