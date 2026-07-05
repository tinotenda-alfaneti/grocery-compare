package compare_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/tinotenda-alfaneti/grocery-compare/internal/compare"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/db"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/giftcard"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/item"
	"github.com/tinotenda-alfaneti/grocery-compare/internal/mapping"
)

// wide-open date window that comfortably contains "today" regardless of when the test runs.
const openFrom, openTo = "2020-01-01", "2099-12-31"

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestCompareItem_ResolvesPromoAndMemberPrice(t *testing.T) {
	conn := openTestDB(t)

	it, err := item.Create(conn, item.CreateInput{Name: "porridge oats"})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	tescoMapping, err := mapping.Create(conn, it.ID, mapping.CreateInput{StoreID: 1, ProductName: "Tesco Porridge Oats 1kg"})
	if err != nil {
		t.Fatalf("create tesco mapping: %v", err)
	}
	aldiMapping, err := mapping.Create(conn, it.ID, mapping.CreateInput{StoreID: 2, ProductName: "Aldi Porridge Oats 1kg"})
	if err != nil {
		t.Fatalf("create aldi mapping: %v", err)
	}

	if err := mapping.AddPrice(conn, tescoMapping.ID, 150, "manual"); err != nil {
		t.Fatalf("add tesco price: %v", err)
	}
	if err := mapping.AddPrice(conn, aldiMapping.ID, 120, "manual"); err != nil {
		t.Fatalf("add aldi price: %v", err)
	}

	// Tesco: promo drops it to 130, but Clubcard price of 99 should win (lower still).
	if _, err := mapping.AddPromo(conn, tescoMapping.ID, mapping.PromoInput{
		PromoPricePence: 130, EffectiveFrom: openFrom, EffectiveTo: openTo,
	}, "manual"); err != nil {
		t.Fatalf("add promo: %v", err)
	}
	if _, err := mapping.AddMemberPrice(conn, tescoMapping.ID, mapping.MemberPriceInput{
		MemberPricePence: 99, EffectiveFrom: openFrom, EffectiveTo: openTo,
	}, "manual"); err != nil {
		t.Fatalf("add member price: %v", err)
	}

	results, err := compare.CompareItem(conn, it.ID)
	if err != nil {
		t.Fatalf("CompareItem: %v", err)
	}

	got := map[int64]compare.ItemComparison{}
	for _, r := range results {
		got[r.StoreID] = r
	}

	if !got[1].Available || got[1].PricePence != 99 {
		t.Fatalf("expected Tesco effective price 99 (Clubcard beats promo), got %+v", got[1])
	}
	if !got[2].Available || got[2].PricePence != 120 {
		t.Fatalf("expected Aldi price 120 (no promo/member), got %+v", got[2])
	}
	if got[3].Available {
		t.Fatalf("expected Asda unavailable (no mapping created), got %+v", got[3])
	}
}

func TestCompareList_AppliesGiftCardDiscountToStoreSubtotal(t *testing.T) {
	conn := openTestDB(t)

	it, err := item.Create(conn, item.CreateInput{Name: "milk 2L"})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	aldiMapping, err := mapping.Create(conn, it.ID, mapping.CreateInput{StoreID: 2, ProductName: "Aldi Milk 2L"})
	if err != nil {
		t.Fatalf("create mapping: %v", err)
	}
	if err := mapping.AddPrice(conn, aldiMapping.ID, 100, "manual"); err != nil {
		t.Fatalf("add price: %v", err)
	}

	if _, err := giftcard.Create(conn, giftcard.CreateInput{
		StoreID: 2, DiscountPercent: 10, EffectiveFrom: openFrom, EffectiveTo: openTo,
	}); err != nil {
		t.Fatalf("add gift card discount: %v", err)
	}

	rec, err := compare.CompareList(conn, []compare.LineItem{{CanonicalItemID: it.ID, Quantity: 5}})
	if err != nil {
		t.Fatalf("CompareList: %v", err)
	}

	var aldiResult *compare.StoreResult
	for i := range rec.PerStore {
		if rec.PerStore[i].Store.ID == 2 {
			aldiResult = &rec.PerStore[i]
		}
	}
	if aldiResult == nil {
		t.Fatalf("expected an Aldi result in %+v", rec.PerStore)
	}
	if aldiResult.SubtotalPence != 500 {
		t.Fatalf("expected subtotal 500 (5 x 100), got %d", aldiResult.SubtotalPence)
	}
	if aldiResult.TotalPence != 450 {
		t.Fatalf("expected total 450 after 10%% gift-card discount, got %d", aldiResult.TotalPence)
	}
}
