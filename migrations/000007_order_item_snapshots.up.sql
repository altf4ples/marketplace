ALTER TABLE store.order_items
    ADD COLUMN product_name VARCHAR(200),
    ADD COLUMN store_name VARCHAR(50);

UPDATE store.order_items oi
SET product_name = p.name,
    store_name = s.name
FROM store.products p
JOIN store.stores s ON s.id = p.store_id
WHERE oi.product_id = p.id
  AND oi.store_id = s.id;

ALTER TABLE store.order_items
    ALTER COLUMN product_name SET NOT NULL,
    ALTER COLUMN store_name SET NOT NULL;

-- Product/store references remain useful for analytics, but historical order
-- records must survive catalog deletion. Keep the IDs without restrictive FKs.
ALTER TABLE store.order_items DROP CONSTRAINT IF EXISTS order_items_product_id_fkey;
ALTER TABLE store.order_items DROP CONSTRAINT IF EXISTS order_items_store_id_fkey;
