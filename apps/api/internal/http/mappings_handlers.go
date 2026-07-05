package httpserver

import (
	"net/http"

	"github.com/tinotenda-alfaneti/grocery-compare/internal/mapping"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/util"
)

func (s *Server) listMappings(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	mappings, err := mapping.ListForItem(s.DB, id)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, mappings)
}

func (s *Server) createMapping(w http.ResponseWriter, r *http.Request) {
	itemID, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	var in mapping.CreateInput
	if err := util.Decode(r, &in); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	created, err := mapping.Create(s.DB, itemID, in)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, created)
}

func (s *Server) updateMapping(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	var in mapping.UpdateInput
	if err := util.Decode(r, &in); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	updated, err := mapping.Update(s.DB, id, in)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, updated)
}

func (s *Server) deleteMapping(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	if err := mapping.Deactivate(s.DB, id); err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) addPrice(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	var body struct {
		PricePence int `json:"pricePence"`
	}
	if err := util.Decode(r, &body); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	if err := mapping.AddPrice(s.DB, id, body.PricePence, "manual"); err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) priceHistory(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	history, err := mapping.PriceHistory(s.DB, id)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, history)
}

func (s *Server) addPromo(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	var in mapping.PromoInput
	if err := util.Decode(r, &in); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	promoID, err := mapping.AddPromo(s.DB, id, in, "manual")
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, map[string]int64{"id": promoID})
}

func (s *Server) deletePromo(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	if err := mapping.DeletePromo(s.DB, id); err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) addMemberPrice(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	var in mapping.MemberPriceInput
	if err := util.Decode(r, &in); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	memberID, err := mapping.AddMemberPrice(s.DB, id, in, "manual")
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, map[string]int64{"id": memberID})
}

func (s *Server) deleteMemberPrice(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	if err := mapping.DeleteMemberPrice(s.DB, id); err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
