# Data model

Full schema: `apps/api/internal/db/schema.sql`. All money is integer pence; all dates are `YYYY-MM-DD` strings.

```
Store
  id, name, slug ('tesco'|'aldi'|'asda'|'lidl')
  supports_member_pricing, member_pricing_label ('Clubcard'|'Rewards'|null)
  included_in_comparisons  -- lets you exclude a store you never visit

CanonicalItem              -- YOUR personal shopping-list vocabulary, e.g. "porridge oats"
  id, name, category, notes, archived

ShoppingList / ShoppingListItem
  lists have a name; items reference a CanonicalItem + quantity + sort_order

ProductMapping              -- "this is the Tesco SKU I buy for 'porridge oats'"
  id, canonical_item_id, store_id
  product_name, product_url (becomes the Phase-2 scrape target), pack_size
  active, is_manual, current_price_pence (cached from the latest PriceObservation)
  UNIQUE(canonical_item_id, store_id) WHERE active

PriceObservation             -- base shelf price, history
  id, mapping_id, price_pence, observed_at, source ('manual'|'scraped')

PromoObservation              -- time-boxed weekly promos
  id, mapping_id, promo_price_pence, promo_label, effective_from, effective_to, source

MemberPriceObservation         -- Clubcard/Rewards personalized price
  id, mapping_id, member_price_pence, effective_from, effective_to (nullable = open-ended), source

GiftCardDiscount               -- employee-perk %, MANUAL ONLY, effective-dated
  id, store_id, discount_percent, effective_from, effective_to, notes

Settings                       -- single row, id=1
  second_stop_min_saving_pence (default 300 = £3.00)
  second_stop_min_saving_percent (optional secondary guard)
  pin_hash, pin_salt
```

## Why no automatic cross-store product matching

Aldi and Lidl stock their own-brand equivalents rather than the same branded SKUs as Tesco/Asda, so there's no reliable automatic 1:1 match across all four stores for most items — that's a real entity-resolution problem, out of scope here. Instead, you maintain your own canonical item ("porridge oats") and map it to whichever specific product you actually buy at each store via `ProductMapping`. Comparisons only ever run over items you've explicitly mapped.

## Known open question: pack-size normalization

`ShoppingListItem.quantity` is one number shared across every store's mapping for that item. If Tesco sells 1kg and Aldi only stocks 500g of your "equivalent" item, `quantity=1` means different real amounts at each store. Phase 1 assumes you map matching pack sizes across stores for a given canonical item (true for most UK staples). A `pack_quantity_multiplier` on `ProductMapping` would be a clean fast-follow if this turns out to matter in practice — deliberately not built preemptively.
