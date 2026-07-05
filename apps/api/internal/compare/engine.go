// Package compare implements the store-comparison and recommendation
// algorithm described in the project plan: given a shopping list, work out
// the cheapest single store and the cheapest worthwhile two-store split.
package compare

import "sort"

type LineItem struct {
	CanonicalItemID int64 `json:"canonicalItemId"`
	Quantity        int   `json:"quantity"`
}

type Store struct {
	ID                      int64   `json:"id"`
	Name                    string  `json:"name"`
	Slug                    string  `json:"slug"`
	SupportsMemberPricing   bool    `json:"supportsMemberPricing"`
	GiftCardDiscountPercent float64 `json:"giftCardDiscountPercent"` // 0 if no active gift-card discount
}

type Settings struct {
	SecondStopMinSavingPence   int      `json:"secondStopMinSavingPence"`
	SecondStopMinSavingPercent *float64 `json:"secondStopMinSavingPercent,omitempty"` // optional secondary guard, nil disables it
}

// PriceLookup returns the effective price (pence) for one unit of a canonical
// item at a given store, and whether that store stocks a mapped product for
// it at all. Callers pass in prices already resolved via EffectivePrice.
type PriceLookup func(storeID, canonicalItemID int64) (pricePence int, ok bool)

type StoreResult struct {
	Store           Store   `json:"store"`
	SubtotalPence   int     `json:"subtotalPence"`
	DiscountPercent float64 `json:"discountPercent"`
	TotalPence      int     `json:"totalPence"`
	CoveredCount    int     `json:"coveredCount"`
	TotalCount      int     `json:"totalCount"`
	CoverageRatio   float64 `json:"coverageRatio"`
}

type SplitAssignment struct {
	CanonicalItemID int64 `json:"canonicalItemId"`
	StoreID         int64 `json:"storeId"`
	PricePence      int   `json:"pricePence"`
}

type SplitResult struct {
	StoreA        Store             `json:"storeA"`
	StoreB        Store             `json:"storeB"`
	TotalAPence   int               `json:"totalAPence"`
	TotalBPence   int               `json:"totalBPence"`
	TotalPence    int               `json:"totalPence"`
	CoverageRatio float64           `json:"coverageRatio"`
	Assignments   []SplitAssignment `json:"assignments"`
}

type Recommendation struct {
	PerStore         []StoreResult `json:"perStore"`
	BestSingle       StoreResult   `json:"bestSingle"`
	BestSplit        *SplitResult  `json:"bestSplit,omitempty"` // nil when fewer than 2 comparable stores
	Recommendation   string        `json:"recommendation"`      // "single" or "split"
	SavingPence      int           `json:"savingPence"`
	SavingPercent    float64       `json:"savingPercent"`
	UnmappedAnywhere []int64       `json:"unmappedAnywhere"`
}

func singleStoreTotal(items []LineItem, store Store, prices PriceLookup) StoreResult {
	subtotal := 0
	covered := 0
	for _, li := range items {
		if p, ok := prices(store.ID, li.CanonicalItemID); ok {
			subtotal += p * li.Quantity
			covered++
		}
	}
	total := subtotal
	if store.GiftCardDiscountPercent > 0 {
		total = subtotal - (subtotal*int(store.GiftCardDiscountPercent*100))/10000
	}
	coverage := 1.0
	if len(items) > 0 {
		coverage = float64(covered) / float64(len(items))
	}
	return StoreResult{
		Store:           store,
		SubtotalPence:   subtotal,
		DiscountPercent: store.GiftCardDiscountPercent,
		TotalPence:      total,
		CoveredCount:    covered,
		TotalCount:      len(items),
		CoverageRatio:   coverage,
	}
}

// bestSingleStore prefers stores that fully cover the list; if none do, it
// falls back to comparing partially-covered stores rather than refusing to
// answer, but the caller can inspect CoverageRatio to warn the user.
func bestSingleStore(items []LineItem, stores []Store, prices PriceLookup) (StoreResult, []StoreResult) {
	results := make([]StoreResult, 0, len(stores))
	for _, s := range stores {
		results = append(results, singleStoreTotal(items, s, prices))
	}

	pool := make([]StoreResult, 0, len(results))
	for _, r := range results {
		if r.CoverageRatio == 1.0 {
			pool = append(pool, r)
		}
	}
	if len(pool) == 0 {
		pool = results
	}

	best := pool[0]
	for _, r := range pool[1:] {
		if r.TotalPence < best.TotalPence {
			best = r
		}
	}
	return best, results
}

func applyDiscount(pence int, percent float64) int {
	if percent <= 0 {
		return pence
	}
	return pence - (pence*int(percent*100))/10000
}

func bestTwoStoreSplit(items []LineItem, stores []Store, prices PriceLookup) *SplitResult {
	if len(stores) < 2 {
		return nil
	}

	var candidates []*SplitResult
	for i := 0; i < len(stores); i++ {
		for j := i + 1; j < len(stores); j++ {
			a, b := stores[i], stores[j]
			subA, subB, covered := 0, 0, 0
			assignments := make([]SplitAssignment, 0, len(items))

			for _, li := range items {
				pA, okA := prices(a.ID, li.CanonicalItemID)
				pB, okB := prices(b.ID, li.CanonicalItemID)
				if !okA && !okB {
					continue
				}
				covered++
				switch {
				case okA && (!okB || pA <= pB):
					subA += pA * li.Quantity
					assignments = append(assignments, SplitAssignment{li.CanonicalItemID, a.ID, pA})
				default:
					subB += pB * li.Quantity
					assignments = append(assignments, SplitAssignment{li.CanonicalItemID, b.ID, pB})
				}
			}

			totalA := applyDiscount(subA, a.GiftCardDiscountPercent)
			totalB := applyDiscount(subB, b.GiftCardDiscountPercent)
			coverage := 1.0
			if len(items) > 0 {
				coverage = float64(covered) / float64(len(items))
			}

			candidates = append(candidates, &SplitResult{
				StoreA:        a,
				StoreB:        b,
				TotalAPence:   totalA,
				TotalBPence:   totalB,
				TotalPence:    totalA + totalB,
				CoverageRatio: coverage,
				Assignments:   assignments,
			})
		}
	}
	if len(candidates) == 0 {
		return nil
	}

	// A pair that stocks nothing costs nothing, which looks "cheapest" by
	// price alone - so, exactly like bestSingleStore, restrict the choice to
	// whichever coverage level the best pairs actually achieve before
	// picking the cheapest among them.
	maxCoverage := 0.0
	for _, c := range candidates {
		if c.CoverageRatio > maxCoverage {
			maxCoverage = c.CoverageRatio
		}
	}

	var best *SplitResult
	for _, c := range candidates {
		if c.CoverageRatio < maxCoverage {
			continue
		}
		if best == nil || c.TotalPence < best.TotalPence {
			best = c
		}
	}
	return best
}

func unmappedAnywhere(items []LineItem, stores []Store, prices PriceLookup) []int64 {
	var out []int64
	for _, li := range items {
		found := false
		for _, s := range stores {
			if _, ok := prices(s.ID, li.CanonicalItemID); ok {
				found = true
				break
			}
		}
		if !found {
			out = append(out, li.CanonicalItemID)
		}
	}
	return out
}

// Recommend computes per-store totals for a shopping list, the cheapest
// single store, the cheapest two-store split, and whether the split saves
// enough to be worth a second trip per Settings.
func Recommend(items []LineItem, stores []Store, prices PriceLookup, settings Settings) Recommendation {
	comparable := make([]Store, 0, len(stores))
	comparable = append(comparable, stores...)
	sort.Slice(comparable, func(i, j int) bool { return comparable[i].ID < comparable[j].ID })

	single, perStore := bestSingleStore(items, comparable, prices)
	split := bestTwoStoreSplit(items, comparable, prices)

	rec := Recommendation{
		PerStore:         perStore,
		BestSingle:       single,
		BestSplit:        split,
		Recommendation:   "single",
		UnmappedAnywhere: unmappedAnywhere(items, comparable, prices),
	}

	if split != nil {
		saving := single.TotalPence - split.TotalPence
		rec.SavingPence = saving
		if single.TotalPence > 0 {
			rec.SavingPercent = 100 * float64(saving) / float64(single.TotalPence)
		}

		worth := saving >= settings.SecondStopMinSavingPence
		if worth && settings.SecondStopMinSavingPercent != nil {
			worth = rec.SavingPercent >= *settings.SecondStopMinSavingPercent
		}
		if worth {
			rec.Recommendation = "split"
		}
	}

	return rec
}
