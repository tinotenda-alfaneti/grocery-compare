package httpserver

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/store"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/util"
)

func (s *Server) listStores(w http.ResponseWriter, r *http.Request) {
	stores, err := store.List(s.DB)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, stores)
}

func (s *Server) updateStore(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	var body struct {
		IncludedInComparisons *bool `json:"includedInComparisons"`
	}
	if err := util.Decode(r, &body); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	if body.IncludedInComparisons != nil {
		if err := store.SetIncluded(s.DB, id, *body.IncludedInComparisons); err != nil {
			util.Err(w, http.StatusInternalServerError, err)
			return
		}
	}
	updated, err := store.Get(s.DB, id)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	if updated == nil {
		util.NotFound(w)
		return
	}
	util.JSON(w, updated)
}
