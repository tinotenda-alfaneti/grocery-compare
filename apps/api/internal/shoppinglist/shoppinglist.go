package shoppinglist

import "database/sql"

type ShoppingList struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Archived  bool   `json:"archived"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type Item struct {
	ID              int64   `json:"id"`
	ShoppingListID  int64   `json:"shoppingListId"`
	CanonicalItemID int64   `json:"canonicalItemId"`
	CanonicalName   string  `json:"canonicalItemName"`
	Quantity        int     `json:"quantity"`
	Notes           *string `json:"notes,omitempty"`
	SortOrder       int     `json:"sortOrder"`
}

func List(db *sql.DB, includeArchived bool) ([]ShoppingList, error) {
	q := `SELECT id, name, archived, created_at, updated_at FROM shopping_lists`
	if !includeArchived {
		q += ` WHERE archived = 0`
	}
	q += ` ORDER BY updated_at DESC`

	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []ShoppingList{}
	for rows.Next() {
		var l ShoppingList
		if err := rows.Scan(&l.ID, &l.Name, &l.Archived, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func Get(db *sql.DB, id int64) (*ShoppingList, error) {
	var l ShoppingList
	err := db.QueryRow(`SELECT id, name, archived, created_at, updated_at FROM shopping_lists WHERE id = ?`, id).
		Scan(&l.ID, &l.Name, &l.Archived, &l.CreatedAt, &l.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func Create(db *sql.DB, name string) (*ShoppingList, error) {
	res, err := db.Exec(`INSERT INTO shopping_lists (name) VALUES (?)`, name)
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
	Archived *bool   `json:"archived,omitempty"`
}

func Update(db *sql.DB, id int64, in UpdateInput) (*ShoppingList, error) {
	if in.Name != nil {
		if _, err := db.Exec(`UPDATE shopping_lists SET name = ?, updated_at = datetime('now') WHERE id = ?`, *in.Name, id); err != nil {
			return nil, err
		}
	}
	if in.Archived != nil {
		if _, err := db.Exec(`UPDATE shopping_lists SET archived = ?, updated_at = datetime('now') WHERE id = ?`, *in.Archived, id); err != nil {
			return nil, err
		}
	}
	return Get(db, id)
}

func Delete(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM shopping_lists WHERE id = ?`, id)
	return err
}

func Items(db *sql.DB, listID int64) ([]Item, error) {
	rows, err := db.Query(`
		SELECT sli.id, sli.shopping_list_id, sli.canonical_item_id, ci.name, sli.quantity, sli.notes, sli.sort_order
		FROM shopping_list_items sli
		JOIN canonical_items ci ON ci.id = sli.canonical_item_id
		WHERE sli.shopping_list_id = ?
		ORDER BY sli.sort_order, sli.id`, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Item{}
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.ShoppingListID, &it.CanonicalItemID, &it.CanonicalName, &it.Quantity, &it.Notes, &it.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

type AddItemInput struct {
	CanonicalItemID int64   `json:"canonicalItemId"`
	Quantity        int     `json:"quantity"`
	Notes           *string `json:"notes,omitempty"`
}

func AddItem(db *sql.DB, listID int64, in AddItemInput) (int64, error) {
	if in.Quantity <= 0 {
		in.Quantity = 1
	}
	var maxOrder sql.NullInt64
	if err := db.QueryRow(`SELECT MAX(sort_order) FROM shopping_list_items WHERE shopping_list_id = ?`, listID).Scan(&maxOrder); err != nil {
		return 0, err
	}
	nextOrder := int64(0)
	if maxOrder.Valid {
		nextOrder = maxOrder.Int64 + 1
	}

	res, err := db.Exec(`
		INSERT INTO shopping_list_items (shopping_list_id, canonical_item_id, quantity, notes, sort_order)
		VALUES (?, ?, ?, ?, ?)`, listID, in.CanonicalItemID, in.Quantity, in.Notes, nextOrder)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

type UpdateItemInput struct {
	Quantity  *int    `json:"quantity,omitempty"`
	Notes     *string `json:"notes,omitempty"`
	SortOrder *int    `json:"sortOrder,omitempty"`
}

func UpdateItem(db *sql.DB, itemID int64, in UpdateItemInput) error {
	if in.Quantity != nil {
		if _, err := db.Exec(`UPDATE shopping_list_items SET quantity = ? WHERE id = ?`, *in.Quantity, itemID); err != nil {
			return err
		}
	}
	if in.Notes != nil {
		if _, err := db.Exec(`UPDATE shopping_list_items SET notes = ? WHERE id = ?`, *in.Notes, itemID); err != nil {
			return err
		}
	}
	if in.SortOrder != nil {
		if _, err := db.Exec(`UPDATE shopping_list_items SET sort_order = ? WHERE id = ?`, *in.SortOrder, itemID); err != nil {
			return err
		}
	}
	return nil
}

func RemoveItem(db *sql.DB, itemID int64) error {
	_, err := db.Exec(`DELETE FROM shopping_list_items WHERE id = ?`, itemID)
	return err
}
