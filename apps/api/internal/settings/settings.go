package settings

import "database/sql"

type Settings struct {
	SecondStopMinSavingPence   int      `json:"secondStopMinSavingPence"`
	SecondStopMinSavingPercent *float64 `json:"secondStopMinSavingPercent,omitempty"`
	PinEnabled                 bool     `json:"pinEnabled"`
}

func Get(db *sql.DB) (*Settings, error) {
	var s Settings
	var pinHash sql.NullString
	err := db.QueryRow(`
		SELECT second_stop_min_saving_pence, second_stop_min_saving_percent, pin_hash
		FROM settings WHERE id = 1`).
		Scan(&s.SecondStopMinSavingPence, &s.SecondStopMinSavingPercent, &pinHash)
	if err != nil {
		return nil, err
	}
	s.PinEnabled = pinHash.Valid && pinHash.String != ""
	return &s, nil
}

type UpdateInput struct {
	SecondStopMinSavingPence   *int     `json:"secondStopMinSavingPence,omitempty"`
	SecondStopMinSavingPercent *float64 `json:"secondStopMinSavingPercent,omitempty"`
}

func Update(db *sql.DB, in UpdateInput) (*Settings, error) {
	if in.SecondStopMinSavingPence != nil {
		if _, err := db.Exec(`UPDATE settings SET second_stop_min_saving_pence = ? WHERE id = 1`, *in.SecondStopMinSavingPence); err != nil {
			return nil, err
		}
	}
	if in.SecondStopMinSavingPercent != nil {
		if _, err := db.Exec(`UPDATE settings SET second_stop_min_saving_percent = ? WHERE id = 1`, *in.SecondStopMinSavingPercent); err != nil {
			return nil, err
		}
	}
	return Get(db)
}
