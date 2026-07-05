// Package auth implements the optional PIN lock: a convenience gate for a
// personal app on a phone, not a real security boundary. No rate limiting,
// no lockout - see docs/architecture.md.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
)

const SessionCookieName = "gc_session"

func generateHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashPin(pin, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + pin))
	return hex.EncodeToString(sum[:])
}

func SetPin(db *sql.DB, pin string) error {
	salt, err := generateHex(16)
	if err != nil {
		return err
	}
	hash := hashPin(pin, salt)
	_, err = db.Exec(`UPDATE settings SET pin_hash = ?, pin_salt = ? WHERE id = 1`, hash, salt)
	return err
}

func VerifyPin(db *sql.DB, pin string) (bool, error) {
	var hash, salt sql.NullString
	if err := db.QueryRow(`SELECT pin_hash, pin_salt FROM settings WHERE id = 1`).Scan(&hash, &salt); err != nil {
		return false, err
	}
	if !hash.Valid || hash.String == "" {
		return false, errors.New("no PIN set")
	}
	candidate := hashPin(pin, salt.String)
	return subtle.ConstantTimeCompare([]byte(candidate), []byte(hash.String)) == 1, nil
}

// sessions is a process-local set of valid session tokens. It resets on
// restart (single replica, personal app) - acceptable for a convenience gate.
var sessions = map[string]bool{}

func NewSession() (string, error) {
	token, err := generateHex(32)
	if err != nil {
		return "", err
	}
	sessions[token] = true
	return token, nil
}

func ValidSession(token string) bool {
	return token != "" && sessions[token]
}
