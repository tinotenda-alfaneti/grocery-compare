package mapping

import (
	"database/sql"
	"fmt"
	"strings"
)

type ProductMapping struct {
	ID                int64   `json:"id"`
	CanonicalItemID   int64   `json:"canonicalItemId"`
	StoreID           int64   `json:"storeId"`
	ProductName       string  `json:"productName"`
	ProductURL        *string `json:"productUrl,omitempty"`
	PackSize          *string `json:"packSize,omitempty"`
	Active            bool    `json:"active"`
	IsManual          bool    `json:"isManual"`
	CurrentPricePence *int    `json:"currentPricePence,omitempty"`
}

type CreateInput struct {
	StoreID     int64   `json:"storeId"`
	ProductName string  `json:"productName"`
	ProductURL  *string `json:"productUrl,omitempty"`
	PackSize    *string `json:"packSize,omitempty"`
}

func ListForItem(db *sql.DB, canonicalItemID int64) ([]ProductMapping, error) {
	rows, err := db.Query(`
		SELECT id, canonical_item_id, store_id, product_name, product_url, pack_size, active, is_manual, current_price_pence
		FROM product_mappings WHERE canonical_item_id = ? AND active = 1 ORDER BY store_id`, canonicalItemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ProductMapping{}
	for rows.Next() {
		var m ProductMapping
		if err := rows.Scan(&m.ID, &m.CanonicalItemID, &m.StoreID, &m.ProductName, &m.ProductURL, &m.PackSize, &m.Active, &m.IsManual, &m.CurrentPricePence); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ListForItems returns all active mappings for a set of canonical items in one query, keyed for cheap lookup.
func ListForItems(db *sql.DB, canonicalItemIDs []int64) ([]ProductMapping, error) {
	if len(canonicalItemIDs) == 0 {
		return nil, nil
	}
	query, args := inClauseQuery(`
		SELECT id, canonical_item_id, store_id, product_name, product_url, pack_size, active, is_manual, current_price_pence
		FROM product_mappings WHERE active = 1 AND canonical_item_id IN (%s)`, canonicalItemIDs)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ProductMapping{}
	for rows.Next() {
		var m ProductMapping
		if err := rows.Scan(&m.ID, &m.CanonicalItemID, &m.StoreID, &m.ProductName, &m.ProductURL, &m.PackSize, &m.Active, &m.IsManual, &m.CurrentPricePence); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func Get(db *sql.DB, id int64) (*ProductMapping, error) {
	var m ProductMapping
	err := db.QueryRow(`
		SELECT id, canonical_item_id, store_id, product_name, product_url, pack_size, active, is_manual, current_price_pence
		FROM product_mappings WHERE id = ?`, id).
		Scan(&m.ID, &m.CanonicalItemID, &m.StoreID, &m.ProductName, &m.ProductURL, &m.PackSize, &m.Active, &m.IsManual, &m.CurrentPricePence)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func Create(db *sql.DB, canonicalItemID int64, in CreateInput) (*ProductMapping, error) {
	res, err := db.Exec(`
		INSERT INTO product_mappings (canonical_item_id, store_id, product_name, product_url, pack_size, active, is_manual)
		VALUES (?, ?, ?, ?, ?, 1, 1)`,
		canonicalItemID, in.StoreID, in.ProductName, in.ProductURL, in.PackSize)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return Get(db, id)
}

type UpdateInput struct {
	ProductName *string `json:"productName,omitempty"`
	ProductURL  *string `json:"productUrl,omitempty"`
	PackSize    *string `json:"packSize,omitempty"`
}

func Update(db *sql.DB, id int64, in UpdateInput) (*ProductMapping, error) {
	if in.ProductName != nil {
		if _, err := db.Exec(`UPDATE product_mappings SET product_name = ? WHERE id = ?`, *in.ProductName, id); err != nil {
			return nil, err
		}
	}
	if in.ProductURL != nil {
		if _, err := db.Exec(`UPDATE product_mappings SET product_url = ? WHERE id = ?`, *in.ProductURL, id); err != nil {
			return nil, err
		}
	}
	if in.PackSize != nil {
		if _, err := db.Exec(`UPDATE product_mappings SET pack_size = ? WHERE id = ?`, *in.PackSize, id); err != nil {
			return nil, err
		}
	}
	return Get(db, id)
}

func Deactivate(db *sql.DB, id int64) error {
	_, err := db.Exec(`UPDATE product_mappings SET active = 0 WHERE id = ?`, id)
	return err
}

// AddPrice records a new observed price and refreshes the mapping's cached current price.
func AddPrice(db *sql.DB, mappingID int64, pricePence int, source string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`INSERT INTO price_observations (mapping_id, price_pence, source) VALUES (?, ?, ?)`, mappingID, pricePence, source); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE product_mappings SET current_price_pence = ? WHERE id = ?`, pricePence, mappingID); err != nil {
		return err
	}
	return tx.Commit()
}

type PriceObservation struct {
	ID         int64  `json:"id"`
	MappingID  int64  `json:"mappingId"`
	PricePence int    `json:"pricePence"`
	ObservedAt string `json:"observedAt"`
	Source     string `json:"source"`
}

func PriceHistory(db *sql.DB, mappingID int64) ([]PriceObservation, error) {
	rows, err := db.Query(`SELECT id, mapping_id, price_pence, observed_at, source FROM price_observations WHERE mapping_id = ? ORDER BY observed_at DESC`, mappingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []PriceObservation{}
	for rows.Next() {
		var p PriceObservation
		if err := rows.Scan(&p.ID, &p.MappingID, &p.PricePence, &p.ObservedAt, &p.Source); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

type PromoInput struct {
	PromoPricePence int     `json:"promoPricePence"`
	PromoLabel      *string `json:"promoLabel,omitempty"`
	EffectiveFrom   string  `json:"effectiveFrom"`
	EffectiveTo     string  `json:"effectiveTo"`
}

func AddPromo(db *sql.DB, mappingID int64, in PromoInput, source string) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO promo_observations (mapping_id, promo_price_pence, promo_label, effective_from, effective_to, source)
		VALUES (?, ?, ?, ?, ?, ?)`,
		mappingID, in.PromoPricePence, in.PromoLabel, in.EffectiveFrom, in.EffectiveTo, source)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func DeletePromo(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM promo_observations WHERE id = ?`, id)
	return err
}

type MemberPriceInput struct {
	MemberPricePence int    `json:"memberPricePence"`
	EffectiveFrom    string `json:"effectiveFrom"`
	EffectiveTo      string `json:"effectiveTo,omitempty"` // empty = open-ended
}

func AddMemberPrice(db *sql.DB, mappingID int64, in MemberPriceInput, source string) (int64, error) {
	var effTo any
	if in.EffectiveTo != "" {
		effTo = in.EffectiveTo
	}
	res, err := db.Exec(`
		INSERT INTO member_price_observations (mapping_id, member_price_pence, effective_from, effective_to, source)
		VALUES (?, ?, ?, ?, ?)`,
		mappingID, in.MemberPricePence, in.EffectiveFrom, effTo, source)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func DeleteMemberPrice(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM member_price_observations WHERE id = ?`, id)
	return err
}

// ActivePromoPrices returns, for the given mapping IDs, the active promo price (pence) on `today` (YYYY-MM-DD), if any.
func ActivePromoPrices(db *sql.DB, mappingIDs []int64, today string) (map[int64]int, error) {
	if len(mappingIDs) == 0 {
		return map[int64]int{}, nil
	}
	query, args := inClauseQuery(`
		SELECT mapping_id, promo_price_pence FROM promo_observations
		WHERE mapping_id IN (%s) AND effective_from <= ? AND effective_to >= ?`, mappingIDs)
	args = append(args, today, today)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int64]int{}
	for rows.Next() {
		var mid int64
		var p int
		if err := rows.Scan(&mid, &p); err != nil {
			return nil, err
		}
		out[mid] = p
	}
	return out, rows.Err()
}

// ActiveMemberPrices returns, for the given mapping IDs, the active Clubcard/Rewards price (pence) on `today`, if any.
func ActiveMemberPrices(db *sql.DB, mappingIDs []int64, today string) (map[int64]int, error) {
	if len(mappingIDs) == 0 {
		return map[int64]int{}, nil
	}
	query, args := inClauseQuery(`
		SELECT mapping_id, member_price_pence FROM member_price_observations
		WHERE mapping_id IN (%s) AND effective_from <= ? AND (effective_to IS NULL OR effective_to >= ?)`, mappingIDs)
	args = append(args, today, today)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int64]int{}
	for rows.Next() {
		var mid int64
		var p int
		if err := rows.Scan(&mid, &p); err != nil {
			return nil, err
		}
		out[mid] = p
	}
	return out, rows.Err()
}

func inClauseQuery(template string, ids []int64) (string, []any) {
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return fmt.Sprintf(template, placeholders), args
}
