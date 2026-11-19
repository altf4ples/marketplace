ALTER TABLE store.order_items
    ADD CONSTRAINT order_items_product_id_fkey
        FOREIGN KEY (product_id) REFERENCES store.products(id),
    ADD CONSTRAINT order_items_store_id_fkey
        FOREIGN KEY (store_id) REFERENCES store.stores(id);

ALTER TABLE store.order_items
    DROP COLUMN IF EXISTS product_name,
    DROP COLUMN IF EXISTS store_name;
