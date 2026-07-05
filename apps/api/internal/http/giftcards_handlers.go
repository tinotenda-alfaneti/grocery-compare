package httpserver

import (
	"net/http"
	"strconv"

	"github.com/tinotenda-alfaneti/grocery-compare/internal/giftcard"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/util"
)

func (s *Server) listGiftCardDiscounts(w http.ResponseWriter, r *http.Request) {
	if storeIDStr := r.URL.Query().Get("storeId"); storeIDStr != "" {
		storeID, err := strconv.ParseInt(storeIDStr, 10, 64)
		if err != nil {
			util.Err(w, http.StatusBadRequest, err)
			return
		}
		discounts, err := giftcard.ListForStore(s.DB, storeID)
		if err != nil {
			util.Err(w, http.StatusInternalServerError, err)
			return
		}
		util.JSON(w, discounts)
		return
	}
	discounts, err := giftcard.ListAll(s.DB)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, discounts)
}

func (s *Server) currentGiftCardDiscounts(w http.ResponseWriter, r *http.Request) {
	byStore, err := giftcard.CurrentByStore(s.DB, todayDate())
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, byStore)
}

func (s *Server) createGiftCardDiscount(w http.ResponseWriter, r *http.Request) {
	var in giftcard.CreateInput
	if err := util.Decode(r, &in); err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	id, err := giftcard.Create(s.DB, in)
	if err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	util.JSON(w, map[string]int64{"id": id})
}

func (s *Server) deleteGiftCardDiscount(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r, "id")
	if err != nil {
		util.Err(w, http.StatusBadRequest, err)
		return
	}
	if err := giftcard.Delete(s.DB, id); err != nil {
		util.Err(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
