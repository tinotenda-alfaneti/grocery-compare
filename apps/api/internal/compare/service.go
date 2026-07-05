// service.go wires the pure algorithm in engine.go to SQLite: it fetches
// mappings/promos/member-prices/gift-card discounts for a shopping list or a
// single item and resolves them into the PriceLookup the engine expects.
package compare

import (
	"database/sql"
	"time"

	"github.com/tinotenda-alfaneti/grocery-compare/internal/giftcard"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/mapping"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/settings"
	storepkg "github.com/tinotenda-alfaneti/grocery-compare/internal/store"
)

func today() string {
	return time.Now().UTC().Format("2006-01-02")
}

// resolvedPrices builds a PriceLookup for the given canonical item IDs by
// fetching their active mappings plus any active promo/member observations
// and resolving each through EffectivePrice.
func resolvedPrices(db *sql.DB, canonicalItemIDs []int64) (PriceLookup, error) {
	mappings, err := mapping.ListForItems(db, canonicalItemIDs)
	if err != nil {
		return nil, err
	}

	mappingIDs := make([]int64, len(mappings))
	storeSupportsMember := map[int64]bool{}
	for i, m := range mappings {
		mappingIDs[i] = m.ID
	}

	stores, err := storepkg.List(db)
	if err != nil {
		return nil, err
	}
	for _, s := range stores {
		storeSupportsMember[s.ID] = s.SupportsMemberPricing
	}

	td := today()
	promos, err := mapping.ActivePromoPrices(db, mappingIDs, td)
	if err != nil {
		return nil, err
	}
	members, err := mapping.ActiveMemberPrices(db, mappingIDs, td)
	if err != nil {
		return nil, err
	}

	// index: storeID -> canonicalItemID -> effective price
	index := map[int64]map[int64]int{}
	for _, m := range mappings {
		if m.CurrentPricePence == nil {
			continue // no price ever recorded for this mapping yet
		}
		var promoPtr, memberPtr *int
		if p, ok := promos[m.ID]; ok {
			promoPtr = &p
		}
		if p, ok := members[m.ID]; ok {
			memberPtr = &p
		}
		price := EffectivePrice(*m.CurrentPricePence, promoPtr, memberPtr, storeSupportsMember[m.StoreID])

		if index[m.StoreID] == nil {
			index[m.StoreID] = map[int64]int{}
		}
		index[m.StoreID][m.CanonicalItemID] = price
	}

	return func(storeID, canonicalItemID int64) (int, bool) {
		byItem, ok := index[storeID]
		if !ok {
			return 0, false
		}
		p, ok := byItem[canonicalItemID]
		return p, ok
	}, nil
}

func loadComparableStores(db *sql.DB) ([]Store, error) {
	all, err := storepkg.List(db)
	if err != nil {
		return nil, err
	}
	discounts, err := giftcard.CurrentByStore(db, today())
	if err != nil {
		return nil, err
	}

	out := make([]Store, 0, len(all))
	for _, s := range all {
		if !s.IncludedInComparisons {
			continue
		}
		out = append(out, Store{
			ID:                      s.ID,
			Name:                    s.Name,
			Slug:                    s.Slug,
			SupportsMemberPricing:   s.SupportsMemberPricing,
			GiftCardDiscountPercent: discounts[s.ID],
		})
	}
	return out, nil
}

func loadSettings(db *sql.DB) (Settings, error) {
	s, err := settings.Get(db)
	if err != nil {
		return Settings{}, err
	}
	return Settings{
		SecondStopMinSavingPence:   s.SecondStopMinSavingPence,
		SecondStopMinSavingPercent: s.SecondStopMinSavingPercent,
	}, nil
}

// CompareList runs the full recommendation for every item currently on a shopping list.
func CompareList(db *sql.DB, lines []LineItem) (Recommendation, error) {
	itemIDs := make([]int64, len(lines))
	for i, l := range lines {
		itemIDs[i] = l.CanonicalItemID
	}

	prices, err := resolvedPrices(db, itemIDs)
	if err != nil {
		return Recommendation{}, err
	}
	stores, err := loadComparableStores(db)
	if err != nil {
		return Recommendation{}, err
	}
	st, err := loadSettings(db)
	if err != nil {
		return Recommendation{}, err
	}

	return Recommend(lines, stores, prices, st), nil
}

// ItemComparison is the per-store price breakdown returned by the single-item lookup endpoint.
type ItemComparison struct {
	StoreID    int64  `json:"storeId"`
	StoreName  string `json:"storeName"`
	Available  bool   `json:"available"`
	PricePence int    `json:"pricePence,omitempty"`
}

// CompareItem looks up a single canonical item's effective price at every comparable store.
func CompareItem(db *sql.DB, canonicalItemID int64) ([]ItemComparison, error) {
	prices, err := resolvedPrices(db, []int64{canonicalItemID})
	if err != nil {
		return nil, err
	}
	stores, err := loadComparableStores(db)
	if err != nil {
		return nil, err
	}

	out := make([]ItemComparison, 0, len(stores))
	for _, s := range stores {
		p, ok := prices(s.ID, canonicalItemID)
		out = append(out, ItemComparison{StoreID: s.ID, StoreName: s.Name, Available: ok, PricePence: p})
	}
	return out, nil
}
