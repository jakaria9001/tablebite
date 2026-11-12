ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS customer_id BIGINT REFERENCES customers(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS currency CHAR(3),
    ADD COLUMN IF NOT EXISTS idempotency_key TEXT,
    ADD COLUMN IF NOT EXISTS accepted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ;

UPDATE orders o
SET currency = r.currency
FROM restaurants r
WHERE o.restaurant_id = r.id
  AND o.currency IS NULL;

UPDATE orders
SET currency = 'INR'
WHERE currency IS NULL;

ALTER TABLE orders
    ALTER COLUMN currency SET DEFAULT 'INR',
    ALTER COLUMN currency SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'orders_currency_format'
    ) THEN
        ALTER TABLE orders
            ADD CONSTRAINT orders_currency_format CHECK (currency ~ '^[A-Z]{3}$');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'orders_subtotal_non_negative'
    ) THEN
        ALTER TABLE orders
            ADD CONSTRAINT orders_subtotal_non_negative CHECK (subtotal >= 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'orders_idempotency_key_length'
    ) THEN
        ALTER TABLE orders
            ADD CONSTRAINT orders_idempotency_key_length
            CHECK (idempotency_key IS NULL OR char_length(idempotency_key) BETWEEN 1 AND 128);
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_customer_idempotency
    ON orders(restaurant_id, customer_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_orders_customer_created
    ON orders(customer_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_orders_table_session_status
    ON orders(table_id, table_session_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_orders_status_created
    ON orders(status, created_at DESC);

ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS item_name_snapshot TEXT,
    ADD COLUMN IF NOT EXISTS selected_variant TEXT,
    ADD COLUMN IF NOT EXISTS unit_price_snapshot NUMERIC(10,2),
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ;

UPDATE order_items
SET item_name_snapshot = item_name
WHERE item_name_snapshot IS NULL;

UPDATE order_items
SET selected_variant = 'regular'
WHERE selected_variant IS NULL;

UPDATE order_items
SET unit_price_snapshot = unit_price
WHERE unit_price_snapshot IS NULL;

UPDATE order_items
SET created_at = NOW()
WHERE created_at IS NULL;

ALTER TABLE order_items
    ALTER COLUMN item_name_snapshot SET NOT NULL,
    ALTER COLUMN selected_variant SET DEFAULT 'regular',
    ALTER COLUMN selected_variant SET NOT NULL,
    ALTER COLUMN unit_price_snapshot SET NOT NULL,
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN created_at SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'order_items_selected_variant_valid'
    ) THEN
        ALTER TABLE order_items
            ADD CONSTRAINT order_items_selected_variant_valid
            CHECK (selected_variant IN ('regular', 'half', 'full'));
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'order_items_quantity_positive'
    ) THEN
        ALTER TABLE order_items
            ADD CONSTRAINT order_items_quantity_positive CHECK (quantity > 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'order_items_unit_price_non_negative'
    ) THEN
        ALTER TABLE order_items
            ADD CONSTRAINT order_items_unit_price_non_negative CHECK (unit_price_snapshot >= 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'order_items_line_total_non_negative'
    ) THEN
        ALTER TABLE order_items
            ADD CONSTRAINT order_items_line_total_non_negative CHECK (line_total >= 0);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_order_items_order
    ON order_items(order_id, id);

CREATE INDEX IF NOT EXISTS idx_order_items_menu_item
    ON order_items(menu_item_id);

ALTER TABLE order_status_history
    ADD COLUMN IF NOT EXISTS previous_status TEXT,
    ADD COLUMN IF NOT EXISTS new_status TEXT,
    ADD COLUMN IF NOT EXISTS changed_by BIGINT REFERENCES admins(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS changed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS note TEXT;

UPDATE order_status_history
SET previous_status = from_status
WHERE previous_status IS NULL;

UPDATE order_status_history
SET new_status = to_status
WHERE new_status IS NULL;

UPDATE order_status_history
SET changed_by = changed_by_admin_id
WHERE changed_by IS NULL;

UPDATE order_status_history
SET changed_at = created_at
WHERE changed_at IS NULL;

ALTER TABLE order_status_history
    ALTER COLUMN new_status SET NOT NULL,
    ALTER COLUMN changed_at SET DEFAULT NOW(),
    ALTER COLUMN changed_at SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'order_status_history_new_status_valid'
    ) THEN
        ALTER TABLE order_status_history
            ADD CONSTRAINT order_status_history_new_status_valid
            CHECK (new_status IN ('new', 'accepted', 'preparing', 'ready', 'served', 'completed', 'cancelled'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_order_status_history_order_changed
    ON order_status_history(order_id, changed_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_order_status_history_changed_by
    ON order_status_history(changed_by, changed_at DESC);
