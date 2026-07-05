package httpserver

import (
	"net/http"
	"time"

	"github.com/tinotenda-alfaneti/grocery-compare/internal/settings"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/util"
)

func todayDate() string {
	return time.Now().UTC().Format("2006-01-02")
}

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	cfg, err := settings.Get(s.DB)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, cfg)
}

func (s *Server) updateSettings(w http.ResponseWriter, r *http.Request) {
	var in settings.UpdateInput
	if err := util.Decode(r, &in); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	updated, err := settings.Update(s.DB, in)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, updated)
}
