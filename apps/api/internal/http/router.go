package httpserver

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	DB *sql.DB
}

func NewRouter(db *sql.DB, webRoot string) http.Handler {
	s := &Server{DB: db}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Route("/api", func(api chi.Router) {
		api.Get("/stores", s.listStores)
		api.Patch("/stores/{id}", s.updateStore)

		api.Get("/items", s.searchItems)
		api.Post("/items", s.createItem)
		api.Get("/items/{id}", s.getItem)
		api.Patch("/items/{id}", s.updateItem)
		api.Delete("/items/{id}", s.archiveItem)
		api.Get("/items/{id}/compare", s.compareItem)

		api.Get("/items/{id}/mappings", s.listMappings)
		api.Post("/items/{id}/mappings", s.createMapping)
		api.Patch("/mappings/{id}", s.updateMapping)
		api.Delete("/mappings/{id}", s.deleteMapping)

		api.Post("/mappings/{id}/price", s.addPrice)
		api.Get("/mappings/{id}/price-history", s.priceHistory)
		api.Post("/mappings/{id}/promo", s.addPromo)
		api.Delete("/promos/{id}", s.deletePromo)
		api.Post("/mappings/{id}/member-price", s.addMemberPrice)
		api.Delete("/member-prices/{id}", s.deleteMemberPrice)

		api.Get("/lists", s.listLists)
		api.Post("/lists", s.createList)
		api.Get("/lists/{id}", s.getList)
		api.Patch("/lists/{id}", s.updateList)
		api.Delete("/lists/{id}", s.deleteList)
		api.Post("/lists/{id}/items", s.addListItem)
		api.Patch("/lists/{id}/items/{itemId}", s.updateListItem)
		api.Delete("/lists/{id}/items/{itemId}", s.removeListItem)
		api.Get("/lists/{id}/compare", s.compareList)

		api.Get("/gift-card-discounts", s.listGiftCardDiscounts)
		api.Get("/gift-card-discounts/current", s.currentGiftCardDiscounts)
		api.Post("/gift-card-discounts", s.createGiftCardDiscount)
		api.Delete("/gift-card-discounts/{id}", s.deleteGiftCardDiscount)

		api.Get("/settings", s.getSettings)
		api.Patch("/settings", s.updateSettings)

		api.Post("/auth/set-pin", s.setPin)
		api.Post("/auth/unlock", s.unlock)
	})

	if webRoot != "" {
		r.NotFound(spaHandler(webRoot))
	}

	return r
}
