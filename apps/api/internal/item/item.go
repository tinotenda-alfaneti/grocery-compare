package item

import (
	"database/sql"
)

type CanonicalItem struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Category *string `json:"category,omitempty"`
	Notes    *string `json:"notes,omitempty"`
	Archived bool    `json:"archived"`
}

type CreateInput struct {
	Name     string  `json:"name"`
	Category *string `json:"category,omitempty"`
	Notes    *string `json:"notes,omitempty"`
}

func Search(db *sql.DB, query string, includeArchived bool) ([]CanonicalItem, error) {
	sqlStr := `SELECT id, name, category, notes, archived FROM canonical_items WHERE 1=1`
	args := []any{}
	if query != "" {
		sqlStr += ` AND name LIKE ?`
		args = append(args, "%"+query+"%")
	}
	if !includeArchived {
		sqlStr += ` AND archived = 0`
	}
	sqlStr += ` ORDER BY name`

	rows, err := db.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []CanonicalItem{}
	for rows.Next() {
		var it CanonicalItem
		if err := rows.Scan(&it.ID, &it.Name, &it.Category, &it.Notes, &it.Archived); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func Get(db *sql.DB, id int64) (*CanonicalItem, error) {
	var it CanonicalItem
	err := db.QueryRow(`SELECT id, name, category, notes, archived FROM canonical_items WHERE id = ?`, id).
		Scan(&it.ID, &it.Name, &it.Category, &it.Notes, &it.Archived)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &it, nil
}

func Create(db *sql.DB, in CreateInput) (*CanonicalItem, error) {
	res, err := db.Exec(`INSERT INTO canonical_items (name, category, notes) VALUES (?, ?, ?)`, in.Name, in.Category, in.Notes)
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
	Name     *string `json:"name,omitempty"`
	Category *string `json:"category,omitempty"`
	Notes    *string `json:"notes,omitempty"`
	Archived *bool   `json:"archived,omitempty"`
}

func Update(db *sql.DB, id int64, in UpdateInput) (*CanonicalItem, error) {
	if in.Name != nil {
		if _, err := db.Exec(`UPDATE canonical_items SET name = ? WHERE id = ?`, *in.Name, id); err != nil {
			return nil, err
		}
	}
	if in.Category != nil {
		if _, err := db.Exec(`UPDATE canonical_items SET category = ? WHERE id = ?`, *in.Category, id); err != nil {
			return nil, err
		}
	}
	if in.Notes != nil {
		if _, err := db.Exec(`UPDATE canonical_items SET notes = ? WHERE id = ?`, *in.Notes, id); err != nil {
			return nil, err
		}
	}
	if in.Archived != nil {
		if _, err := db.Exec(`UPDATE canonical_items SET archived = ? WHERE id = ?`, *in.Archived, id); err != nil {
			return nil, err
		}
	}
	return Get(db, id)
}

// Archive soft-deletes an item rather than removing its history.
func Archive(db *sql.DB, id int64) error {
	_, err := db.Exec(`UPDATE canonical_items SET archived = 1 WHERE id = ?`, id)
	return err
}
