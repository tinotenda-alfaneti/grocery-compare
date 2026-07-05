package httpserver

import (
	"net/http"

	"github.com/tinotenda-alfaneti/grocery-compare/internal/compare"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/shoppinglist"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/util"
)

func (s *Server) listLists(w http.ResponseWriter, r *http.Request) {
	lists, err := shoppinglist.List(s.DB, false)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, lists)
}

func (s *Server) createList(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := util.Decode(r, &body); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	created, err := shoppinglist.Create(s.DB, body.Name)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, created)
}

func (s *Server) getList(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	list, err := shoppinglist.Get(s.DB, id)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	if list == nil {
		util.NotFound(w)
		return
	}
	items, err := shoppinglist.Items(s.DB, id)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, map[string]any{"list": list, "items": items})
}

func (s *Server) updateList(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	var in shoppinglist.UpdateInput
	if err := util.Decode(r, &in); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	updated, err := shoppinglist.Update(s.DB, id, in)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, updated)
}

func (s *Server) deleteList(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	if err := shoppinglist.Delete(s.DB, id); err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) addListItem(w http.ResponseWriter, r *http.Request) {
	listID, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	var in shoppinglist.AddItemInput
	if err := util.Decode(r, &in); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	itemID, err := shoppinglist.AddItem(s.DB, listID, in)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, map[string]int64{"id": itemID})
}

func (s *Server) updateListItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := idParam(r, "itemId")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	var in shoppinglist.UpdateItemInput
	if err := util.Decode(r, &in); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	if err := shoppinglist.UpdateItem(s.DB, itemID, in); err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) removeListItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := idParam(r, "itemId")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	if err := shoppinglist.RemoveItem(s.DB, itemID); err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) compareList(w http.ResponseWriter, r *http.Request) {
	listID, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	items, err := shoppinglist.Items(s.DB, listID)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	lines := make([]compare.LineItem, len(items))
	for i, it := range items {
		lines[i] = compare.LineItem{CanonicalItemID: it.CanonicalItemID, Quantity: it.Quantity}
	}
	result, err := compare.CompareList(s.DB, lines)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, result)
}
