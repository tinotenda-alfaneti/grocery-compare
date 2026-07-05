package compare

import "testing"

func store(id int64, name string, discount float64) Store {
	return Store{ID: id, Name: name, GiftCardDiscountPercent: discount}
}

// fixedPrices builds a PriceLookup from a nested map: storeID -> itemID -> price.
func fixedPrices(m map[int64]map[int64]int) PriceLookup {
	return func(storeID, itemID int64) (int, bool) {
		byItem, ok := m[storeID]
		if !ok {
			return 0, false
		}
		p, ok := byItem[itemID]
		return p, ok
	}
}

func TestSingleStoreTotal_FullCoverage(t *testing.T) {
	items := []LineItem{{CanonicalItemID: 1, Quantity: 2}, {CanonicalItemID: 2, Quantity: 1}}
	prices := fixedPrices(map[int64]map[int64]int{
		10: {1: 100, 2: 250},
	})
	res := singleStoreTotal(items, store(10, "Tesco", 0), prices)

	if res.SubtotalPence != 450 {
		t.Fatalf("expected subtotal 450, got %d", res.SubtotalPence)
	}
	if res.TotalPence != 450 {
		t.Fatalf("expected total 450 (no discount), got %d", res.TotalPence)
	}
	if res.CoverageRatio != 1.0 {
		t.Fatalf("expected full coverage, got %v", res.CoverageRatio)
	}
}

func TestSingleStoreTotal_PartialCoverage(t *testing.T) {
	items := []LineItem{{CanonicalItemID: 1, Quantity: 1}, {CanonicalItemID: 2, Quantity: 1}}
	prices := fixedPrices(map[int64]map[int64]int{
		10: {1: 100}, // item 2 not stocked here
	})
	res := singleStoreTotal(items, store(10, "Tesco", 0), prices)

	if res.CoveredCount != 1 || res.TotalCount != 2 {
		t.Fatalf("expected 1/2 covered, got %d/%d", res.CoveredCount, res.TotalCount)
	}
	if res.CoverageRatio != 0.5 {
		t.Fatalf("expected 0.5 coverage, got %v", res.CoverageRatio)
	}
	if res.SubtotalPence != 100 {
		t.Fatalf("expected subtotal 100, got %d", res.SubtotalPence)
	}
}

func TestSingleStoreTotal_DiscountAppliedOncePerSubtotalNotPerItem(t *testing.T) {
	items := []LineItem{{CanonicalItemID: 1, Quantity: 3}, {CanonicalItemID: 2, Quantity: 1}}
	prices := fixedPrices(map[int64]map[int64]int{
		10: {1: 100, 2: 100}, // subtotal = 400
	})
	res := singleStoreTotal(items, store(10, "Tesco", 10), prices) // 10% off

	if res.SubtotalPence != 400 {
		t.Fatalf("expected subtotal 400, got %d", res.SubtotalPence)
	}
	// 10% of 400 = 40 off -> 360, computed once on the summed subtotal.
	if res.TotalPence != 360 {
		t.Fatalf("expected total 360 after 10%% off summed subtotal, got %d", res.TotalPence)
	}
}

func TestBestSingleStore_PrefersFullCoverageOverCheaperPartial(t *testing.T) {
	items := []LineItem{{CanonicalItemID: 1, Quantity: 1}, {CanonicalItemID: 2, Quantity: 1}}
	prices := fixedPrices(map[int64]map[int64]int{
		10: {1: 100, 2: 100}, // Tesco: full coverage, total 200
		20: {1: 10},          // Aldi: cheaper but only covers 1 of 2 items
	})
	best, _ := bestSingleStore(items, []Store{store(10, "Tesco", 0), store(20, "Aldi", 0)}, prices)

	if best.Store.ID != 10 {
		t.Fatalf("expected fully-covered Tesco to win despite higher price, got store %d", best.Store.ID)
	}
}

func TestBestSingleStore_FallsBackToPartialWhenNoStoreFullyCovers(t *testing.T) {
	items := []LineItem{{CanonicalItemID: 1, Quantity: 1}, {CanonicalItemID: 2, Quantity: 1}}
	prices := fixedPrices(map[int64]map[int64]int{
		10: {1: 100}, // covers item 1 only
		20: {2: 50},  // covers item 2 only
	})
	best, all := bestSingleStore(items, []Store{store(10, "Tesco", 0), store(20, "Aldi", 0)}, prices)

	if best.CoverageRatio == 1.0 {
		t.Fatalf("no store should report full coverage in this fixture")
	}
	if len(all) != 2 {
		t.Fatalf("expected results for both stores, got %d", len(all))
	}
}

// TestBestTwoStoreSplit_IgnoresZeroCoveragePairs guards against a real bug
// found via manual end-to-end testing: a pair of stores that stock nothing
// on the list costs nothing, which looked like the "cheapest" split even
// though it's a useless recommendation (you'd come home empty-handed).
func TestBestTwoStoreSplit_IgnoresZeroCoveragePairs(t *testing.T) {
	items := []LineItem{{CanonicalItemID: 1, Quantity: 1}}
	prices := fixedPrices(map[int64]map[int64]int{
		10: {1: 120}, // Tesco stocks it
		20: {1: 150}, // Aldi stocks it too
		// Asda (30) and Lidl (40) stock nothing for this item.
	})
	stores := []Store{store(10, "Tesco", 0), store(20, "Aldi", 0), store(30, "Asda", 0), store(40, "Lidl", 0)}

	split := bestTwoStoreSplit(items, stores, prices)
	if split == nil {
		t.Fatalf("expected a split result")
	}
	if split.CoverageRatio != 1.0 {
		t.Fatalf("expected the chosen split to fully cover the list, got coverage %v (stores %s/%s)",
			split.CoverageRatio, split.StoreA.Name, split.StoreB.Name)
	}
	if split.TotalPence != 120 {
		t.Fatalf("expected the cheapest fully-covering split to cost 120 (Tesco), got %d", split.TotalPence)
	}
}

func TestUnmappedAnywhere(t *testing.T) {
	items := []LineItem{{CanonicalItemID: 1, Quantity: 1}, {CanonicalItemID: 99, Quantity: 1}}
	prices := fixedPrices(map[int64]map[int64]int{
		10: {1: 100},
	})
	missing := unmappedAnywhere(items, []Store{store(10, "Tesco", 0)}, prices)

	if len(missing) != 1 || missing[0] != 99 {
		t.Fatalf("expected item 99 reported unmapped, got %v", missing)
	}
}

// TestRecommend_SplitWorthIt is one of the two hand-checkable fixtures called
// for in the verification plan: a basket where splitting across two stores
// saves more than the configured threshold.
func TestRecommend_SplitWorthIt(t *testing.T) {
	items := []LineItem{
		{CanonicalItemID: 1, Quantity: 1}, // cheap at Aldi
		{CanonicalItemID: 2, Quantity: 1}, // cheap at Lidl
	}
	prices := fixedPrices(map[int64]map[int64]int{
		10: {1: 500, 2: 500}, // Tesco: 1000 total, fully covered
		20: {1: 100, 2: 900}, // Aldi
		30: {1: 900, 2: 100}, // Lidl
	})
	stores := []Store{store(10, "Tesco", 0), store(20, "Aldi", 0), store(30, "Lidl", 0)}
	settings := Settings{SecondStopMinSavingPence: 300}

	rec := Recommend(items, stores, prices, settings)

	// Best single store is Tesco at 1000 (only fully-covered store; Aldi/Lidl
	// each cover both items too at 100+900=1000, so all three are equal here
	// for the single-store case; use bigger contrast below for stronger signal).
	if rec.BestSplit == nil {
		t.Fatalf("expected a split result with >=2 stores")
	}
	wantSplitTotal := 200 // 100 (item1 @ Aldi) + 100 (item2 @ Lidl)
	if rec.BestSplit.TotalPence != wantSplitTotal {
		t.Fatalf("expected split total %d, got %d", wantSplitTotal, rec.BestSplit.TotalPence)
	}
	if rec.SavingPence < 300 {
		t.Fatalf("expected saving >= threshold 300, got %d", rec.SavingPence)
	}
	if rec.Recommendation != "split" {
		t.Fatalf("expected recommendation 'split', got %q", rec.Recommendation)
	}
}

// TestRecommend_SplitNotWorthIt mirrors the fixture above but shrinks the
// saving below the configured threshold, and expects the engine to recommend
// staying at a single store instead.
func TestRecommend_SplitNotWorthIt(t *testing.T) {
	items := []LineItem{
		{CanonicalItemID: 1, Quantity: 1},
		{CanonicalItemID: 2, Quantity: 1},
	}
	prices := fixedPrices(map[int64]map[int64]int{
		10: {1: 500, 2: 501}, // Tesco: 1001 total
		20: {1: 495, 2: 501}, // Aldi: 996 total, cheapest single store
		30: {1: 500, 2: 496}, // Lidl
	})
	stores := []Store{store(10, "Tesco", 0), store(20, "Aldi", 0), store(30, "Lidl", 0)}
	settings := Settings{SecondStopMinSavingPence: 300}

	rec := Recommend(items, stores, prices, settings)

	if rec.SavingPence >= 300 {
		t.Fatalf("fixture should produce a saving under the 300p threshold, got %d", rec.SavingPence)
	}
	if rec.Recommendation != "single" {
		t.Fatalf("expected recommendation 'single' when saving is below threshold, got %q", rec.Recommendation)
	}
}

func TestRecommend_ExactlyAtThresholdCountsAsWorthIt(t *testing.T) {
	items := []LineItem{{CanonicalItemID: 1, Quantity: 1}, {CanonicalItemID: 2, Quantity: 1}}
	prices := fixedPrices(map[int64]map[int64]int{
		10: {1: 700, 2: 700}, // Tesco: 1400 total, deliberately uncompetitive
		20: {1: 350, 2: 650}, // Aldi: 1000 total; cheapest for item1
		30: {1: 650, 2: 350}, // Lidl: 1000 total; cheapest for item2
	})
	stores := []Store{store(10, "Tesco", 0), store(20, "Aldi", 0), store(30, "Lidl", 0)}
	settings := Settings{SecondStopMinSavingPence: 300}

	rec := Recommend(items, stores, prices, settings)

	// Best single store: Aldi or Lidl at 1000. Best split: 350 (item1@Aldi) + 350 (item2@Lidl) = 700.
	// Saving = 1000 - 700 = 300 exactly.
	if rec.SavingPence != 300 {
		t.Fatalf("expected saving exactly 300, got %d", rec.SavingPence)
	}
	if rec.Recommendation != "split" {
		t.Fatalf("expected exactly-at-threshold saving to count as worth it, got %q", rec.Recommendation)
	}
}

func TestIsActive(t *testing.T) {
	cases := []struct {
		from, to, today string
		want            bool
	}{
		{"2026-01-01", "2026-01-31", "2026-01-15", true},
		{"2026-01-01", "2026-01-31", "2026-02-01", false},
		{"2026-01-01", "2026-01-31", "2025-12-31", false},
		{"2026-01-01", "", "2030-01-01", true}, // open-ended
	}
	for _, c := range cases {
		if got := IsActive(c.from, c.to, c.today); got != c.want {
			t.Errorf("IsActive(%q,%q,%q) = %v, want %v", c.from, c.to, c.today, got, c.want)
		}
	}
}

func TestEffectivePrice(t *testing.T) {
	promo := 80
	member := 70
	higherMember := 95

	if got := EffectivePrice(100, nil, nil, true); got != 100 {
		t.Errorf("no promo/member: expected 100, got %d", got)
	}
	if got := EffectivePrice(100, &promo, nil, true); got != 80 {
		t.Errorf("promo only: expected 80, got %d", got)
	}
	if got := EffectivePrice(100, &promo, &member, true); got != 70 {
		t.Errorf("promo+member: expected min 70, got %d", got)
	}
	if got := EffectivePrice(100, nil, &higherMember, true); got != 95 {
		t.Errorf("member higher than shelf still applies (guards bad data, not clamped): expected 95, got %d", got)
	}
	if got := EffectivePrice(100, nil, &member, false); got != 100 {
		t.Errorf("member price ignored when store doesn't support it: expected 100, got %d", got)
	}
}
