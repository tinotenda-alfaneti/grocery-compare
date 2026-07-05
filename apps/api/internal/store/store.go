package store

import "database/sql"

type Store struct {
	ID                    int64   `json:"id"`
	Name                  string  `json:"name"`
	Slug                  string  `json:"slug"`
	SupportsMemberPricing bool    `json:"supportsMemberPricing"`
	MemberPricingLabel    *string `json:"memberPricingLabel,omitempty"`
	IncludedInComparisons bool    `json:"includedInComparisons"`
}

func List(db *sql.DB) ([]Store, error) {
	rows, err := db.Query(`SELECT id, name, slug, supports_member_pricing, member_pricing_label, included_in_comparisons FROM stores ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Store
	for rows.Next() {
		var s Store
		if err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.SupportsMemberPricing, &s.MemberPricingLabel, &s.IncludedInComparisons); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func Get(db *sql.DB, id int64) (*Store, error) {
	var s Store
	err := db.QueryRow(`SELECT id, name, slug, supports_member_pricing, member_pricing_label, included_in_comparisons FROM stores WHERE id = ?`, id).
		Scan(&s.ID, &s.Name, &s.Slug, &s.SupportsMemberPricing, &s.MemberPricingLabel, &s.IncludedInComparisons)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func SetIncluded(db *sql.DB, id int64, included bool) error {
	_, err := db.Exec(`UPDATE stores SET included_in_comparisons = ? WHERE id = ?`, included, id)
	return err
}
