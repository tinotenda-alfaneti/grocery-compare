package giftcard

import "database/sql"

type Discount struct {
	ID              int64   `json:"id"`
	StoreID         int64   `json:"storeId"`
	DiscountPercent float64 `json:"discountPercent"`
	EffectiveFrom   string  `json:"effectiveFrom"`
	EffectiveTo     *string `json:"effectiveTo,omitempty"`
	Notes           *string `json:"notes,omitempty"`
}

type CreateInput struct {
	StoreID         int64   `json:"storeId"`
	DiscountPercent float64 `json:"discountPercent"`
	EffectiveFrom   string  `json:"effectiveFrom"`
	EffectiveTo     string  `json:"effectiveTo,omitempty"` // empty = open-ended
	Notes           *string `json:"notes,omitempty"`
}

func Create(db *sql.DB, in CreateInput) (int64, error) {
	var effTo any
	if in.EffectiveTo != "" {
		effTo = in.EffectiveTo
	}
	res, err := db.Exec(`
		INSERT INTO gift_card_discounts (store_id, discount_percent, effective_from, effective_to, notes)
		VALUES (?, ?, ?, ?, ?)`,
		in.StoreID, in.DiscountPercent, in.EffectiveFrom, effTo, in.Notes)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ListForStore(db *sql.DB, storeID int64) ([]Discount, error) {
	rows, err := db.Query(`
		SELECT id, store_id, discount_percent, effective_from, effective_to, notes
		FROM gift_card_discounts WHERE store_id = ? ORDER BY effective_from DESC`, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAll(rows)
}

func ListAll(db *sql.DB) ([]Discount, error) {
	rows, err := db.Query(`
		SELECT id, store_id, discount_percent, effective_from, effective_to, notes
		FROM gift_card_discounts ORDER BY store_id, effective_from DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAll(rows)
}

func scanAll(rows *sql.Rows) ([]Discount, error) {
	out := []Discount{}
	for rows.Next() {
		var d Discount
		if err := rows.Scan(&d.ID, &d.StoreID, &d.DiscountPercent, &d.EffectiveFrom, &d.EffectiveTo, &d.Notes); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func Delete(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM gift_card_discounts WHERE id = ?`, id)
	return err
}

// CurrentByStore returns the active discount percent (0 if none) for every store, keyed by store ID, as of `today`.
// When multiple discounts are active for a store (shouldn't normally happen), the most recently started one wins.
func CurrentByStore(db *sql.DB, today string) (map[int64]float64, error) {
	rows, err := db.Query(`
		SELECT store_id, discount_percent FROM gift_card_discounts
		WHERE effective_from <= ? AND (effective_to IS NULL OR effective_to >= ?)
		ORDER BY store_id, effective_from DESC`, today, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int64]float64{}
	for rows.Next() {
		var storeID int64
		var pct float64
		if err := rows.Scan(&storeID, &pct); err != nil {
			return nil, err
		}
		if _, seen := out[storeID]; !seen {
			out[storeID] = pct
		}
	}
	return out, rows.Err()
}
