CREATE TABLE store.notifications (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID       NOT NULL REFERENCES store.users(id)  ON DELETE CASCADE,
    order_id    UUID       REFERENCES store.orders(id)          ON DELETE CASCADE,
    message     TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user_id ON store.notifications(user_id);