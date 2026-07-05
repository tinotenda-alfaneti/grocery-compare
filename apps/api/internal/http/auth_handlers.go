package httpserver

import (
	"errors"
	"net/http"

	"github.com/tinotenda-alfaneti/grocery-compare/internal/auth"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/util"
)

func (s *Server) setPin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Pin string `json:"pin"`
	}
	if err := util.Decode(r, &body); err != nil || body.Pin == "" {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	if err := auth.SetPin(s.DB, body.Pin); err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) unlock(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Pin string `json:"pin"`
	}
	if err := util.Decode(r, &body); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	ok, err := auth.VerifyPin(s.DB, body.Pin)
	if err != nil || !ok {
		util.Err(w, http.StatusUnauthorized, errors.New("incorrect PIN"))
		return
	}
	token, err := auth.NewSession()
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

