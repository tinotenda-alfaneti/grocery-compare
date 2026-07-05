package httpserver

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/compare"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/item"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/util"
)

func idParam(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, name), 10, 64)
}

func (s *Server) searchItems(w http.ResponseWriter, r *http.Request) {
	items, err := item.Search(s.DB, r.URL.Query().Get("query"), false)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, items)
}

func (s *Server) createItem(w http.ResponseWriter, r *http.Request) {
	var in item.CreateInput
	if err := util.Decode(r, &in); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	created, err := item.Create(s.DB, in)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, created)
}

func (s *Server) getItem(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	it, err := item.Get(s.DB, id)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	if it == nil {
		util.NotFound(w)
		return
	}
	util.JSON(w, it)
}

func (s *Server) updateItem(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	var in item.UpdateInput
	if err := util.Decode(r, &in); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	updated, err := item.Update(s.DB, id, in)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, updated)
}

func (s *Server) archiveItem(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	if err := item.Archive(s.DB, id); err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) compareItem(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	result, err := compare.CompareItem(s.DB, id)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, result)
}
