CREATE TABLE IF NOT EXISTS stores (
    id                      INTEGER PRIMARY KEY,
    name                    TEXT NOT NULL,
    slug                    TEXT NOT NULL UNIQUE,
    supports_member_pricing INTEGER NOT NULL DEFAULT 0,
    member_pricing_label    TEXT,
    included_in_comparisons INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS canonical_items (
    id         INTEGER PRIMARY KEY,
    name       TEXT NOT NULL,
    category   TEXT,
    notes      TEXT,
    archived   INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_canonical_items_name ON canonical_items(name);

CREATE TABLE IF NOT EXISTS shopping_lists (
    id         INTEGER PRIMARY KEY,
    name       TEXT NOT NULL,
    archived   INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS shopping_list_items (
    id                INTEGER PRIMARY KEY,
    shopping_list_id  INTEGER NOT NULL REFERENCES shopping_lists(id) ON DELETE CASCADE,
    canonical_item_id INTEGER NOT NULL REFERENCES canonical_items(id),
    quantity          INTEGER NOT NULL DEFAULT 1,
    notes             TEXT,
    sort_order        INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_shopping_list_items_list ON shopping_list_items(shopping_list_id);

CREATE TABLE IF NOT EXISTS product_mappings (
    id                  INTEGER PRIMARY KEY,
    canonical_item_id   INTEGER NOT NULL REFERENCES canonical_items(id),
    store_id            INTEGER NOT NULL REFERENCES stores(id),
    product_name        TEXT NOT NULL,
    product_url         TEXT,
    pack_size           TEXT,
    active              INTEGER NOT NULL DEFAULT 1,
    is_manual           INTEGER NOT NULL DEFAULT 1,
    current_price_pence INTEGER,
    created_at          TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_product_mappings_active_unique
    ON product_mappings(canonical_item_id, store_id) WHERE active = 1;
CREATE INDEX IF NOT EXISTS idx_product_mappings_item ON product_mappings(canonical_item_id);

CREATE TABLE IF NOT EXISTS price_observations (
    id          INTEGER PRIMARY KEY,
    mapping_id  INTEGER NOT NULL REFERENCES product_mappings(id) ON DELETE CASCADE,
    price_pence INTEGER NOT NULL,
    observed_at TEXT NOT NULL DEFAULT (datetime('now')),
    source      TEXT NOT NULL DEFAULT 'manual'
);
CREATE INDEX IF NOT EXISTS idx_price_observations_mapping ON price_observations(mapping_id, observed_at);

CREATE TABLE IF NOT EXISTS promo_observations (
    id                INTEGER PRIMARY KEY,
    mapping_id        INTEGER NOT NULL REFERENCES product_mappings(id) ON DELETE CASCADE,
    promo_price_pence INTEGER NOT NULL,
    promo_label       TEXT,
    effective_from    TEXT NOT NULL,
    effective_to      TEXT NOT NULL,
    source            TEXT NOT NULL DEFAULT 'manual'
);
CREATE INDEX IF NOT EXISTS idx_promo_observations_mapping ON promo_observations(mapping_id, effective_from, effective_to);

CREATE TABLE IF NOT EXISTS member_price_observations (
    id                 INTEGER PRIMARY KEY,
    mapping_id         INTEGER NOT NULL REFERENCES product_mappings(id) ON DELETE CASCADE,
    member_price_pence INTEGER NOT NULL,
    effective_from     TEXT NOT NULL,
    effective_to       TEXT,
    source             TEXT NOT NULL DEFAULT 'manual'
);
CREATE INDEX IF NOT EXISTS idx_member_price_observations_mapping ON member_price_observations(mapping_id, effective_from);

CREATE TABLE IF NOT EXISTS gift_card_discounts (
    id               INTEGER PRIMARY KEY,
    store_id         INTEGER NOT NULL REFERENCES stores(id),
    discount_percent REAL NOT NULL,
    effective_from   TEXT NOT NULL,
    effective_to     TEXT,
    notes            TEXT,
    created_at       TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_gift_card_discounts_store ON gift_card_discounts(store_id, effective_from);

CREATE TABLE IF NOT EXISTS settings (
    id                              INTEGER PRIMARY KEY CHECK (id = 1),
    second_stop_min_saving_pence   INTEGER NOT NULL DEFAULT 300,
    second_stop_min_saving_percent REAL,
    pin_hash                        TEXT,
    pin_salt                        TEXT
);

INSERT OR IGNORE INTO stores (id, name, slug, supports_member_pricing, member_pricing_label, included_in_comparisons)
VALUES
    (1, 'Tesco', 'tesco', 1, 'Clubcard', 1),
    (2, 'Aldi',  'aldi',  0, NULL,       1),
    (3, 'Asda',  'asda',  1, 'Rewards',  1),
    (4, 'Lidl',  'lidl',  0, NULL,       1);

INSERT OR IGNORE INTO settings (id, second_stop_min_saving_pence, second_stop_min_saving_percent)
VALUES (1, 300, NULL);
