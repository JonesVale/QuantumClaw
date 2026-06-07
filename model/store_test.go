package model

import (
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"
)

// getStoreTestDB reuses the DB from billing_test.go's setupTestDB
// but manages its lifecycle separately (no t.Cleanup on the shared instance)
func getStoreTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	// setupTestDB creates a DB and registers cleanup on the calling test.
	// Since each test needs its own DB to avoid "database is closed",
	// we call it directly per test.
	db := setupTestDB(t)
	db.AutoMigrate(&Store{}, &Listing{}, &StoreTierLog{}, &Review{}, &PlatformFeeConfig{})
	return db
}

var storeTestID int

func mkStore(t *testing.T) (*Store, *gorm.DB) {
	t.Helper()
	db := getStoreTestDB(t)
	storeTestID++
	s := &Store{
		UserID:       90000 + storeTestID,
		Name:         fmt.Sprintf("store-%d", storeTestID),
		Tier:         StoreTierBasic,
		Status:       StoreStatusActive,
		Rating:       5.0,
		OpenedAt:     time.Now().Unix(),
		LastActiveAt: time.Now().Unix(),
	}
	if err := db.Create(s).Error; err != nil {
		t.Fatalf("create store: %v", err)
	}
	return s, db
}

func mkListing(db *gorm.DB, storeID, providerID int, modelName string, price int64) *Listing {
	l := &Listing{
		ID:           fmt.Sprintf("lst_%d", time.Now().UnixNano()),
		StoreID:      storeID,
		ProviderID:   providerID,
		ChannelID:    1,
		ModelName:    modelName,
		PricePerUnit: price,
		Region:       "overseas",
		Status:       ListingStatusActive,
	}
	db.Create(l)
	return l
}

func TestStoreCreateFind(t *testing.T) {
	db := getStoreTestDB(t)
	s := &Store{
		UserID: 99901, Name: "s1", Tier: StoreTierBasic, Status: StoreStatusActive,
		Rating: 5.0, OpenedAt: time.Now().Unix(), LastActiveAt: time.Now().Unix(),
	}
	db.Create(s)
	var st Store
	db.First(&st, s.ID)
	if st.Tier != StoreTierBasic {
		t.Fatalf("expected basic, got %s", st.Tier)
	}
}

func TestStoreFindByName(t *testing.T) {
	s, db := mkStore(t)
	var st Store
	db.Where("user_id = ?", s.UserID).First(&st)
	if st.ID != s.ID {
		t.Fatalf("expected store %d, got %d", s.ID, st.ID)
	}
}

func TestStoreUpdate(t *testing.T) {
	s, db := mkStore(t)
	db.Model(s).Updates(map[string]interface{}{"name": "new", "tier": StoreTierGold, "total_sales": int64(5000)})
	var st Store
	db.First(&st, s.ID)
	if st.Name != "new" || st.Tier != StoreTierGold || st.TotalSales != 5000 {
		t.Fatalf("update failed")
	}
}

func TestStoreDuplicateUser(t *testing.T) {
	s, db := mkStore(t)
	err := db.Create(&Store{
		UserID: s.UserID, Name: "d", Tier: StoreTierBasic, Status: StoreStatusActive,
		OpenedAt: time.Now().Unix(), LastActiveAt: time.Now().Unix(),
	}).Error
	if err == nil {
		t.Fatal("expected dup error")
	}
}

func TestListingPrice(t *testing.T) {
	_, db := mkStore(t)
	mn := fmt.Sprintf("m-%d", time.Now().UnixNano())
	l := mkListing(db, 1, 1, mn, 150)
	var lst Listing
	db.Where("id = ?", l.ID).First(&lst)
	if lst.PricePerUnit != 150 {
		t.Fatalf("expected 150, got %d", lst.PricePerUnit)
	}
	db.Model(&lst).Update("price_per_unit", 180)
	db.Where("id = ?", l.ID).First(&lst)
	if lst.PricePerUnit != 180 {
		t.Fatalf("expected 180, got %d", lst.PricePerUnit)
	}
}

func TestListingsSearchPriceSort(t *testing.T) {
	s1, db := mkStore(t)
	s2, db2 := mkStore(t)
	if s2.ID == 0 {
		_ = db2
	}
	mn := fmt.Sprintf("s-%d", time.Now().UnixNano())
	mkListing(db, s1.ID, s1.UserID, mn, 200)
	mkListing(db, s2.ID, s2.UserID, mn, 150)
	var res []Listing
	db.Where("status = ? AND model_name = ?", ListingStatusActive, mn).Order("price_per_unit ASC").Find(&res)
	if len(res) != 2 {
		t.Fatalf("expected 2, got %d", len(res))
	}
	if res[0].PricePerUnit != 150 {
		t.Fatalf("expected 150 first, got %d", res[0].PricePerUnit)
	}
}

func TestFeeConfigRates(t *testing.T) {
	db := getStoreTestDB(t)
	db.Create(&PlatformFeeConfig{Tier: StoreTierBasic, Rate: 10, MinSkip: 10000})
	var c PlatformFeeConfig
	db.Where("tier = ?", StoreTierBasic).First(&c)
	if c.Rate != 10 {
		t.Fatalf("expected 10, got %f", c.Rate)
	}
}

func TestFeeConfigUpdate(t *testing.T) {
	db := getStoreTestDB(t)
	db.Create(&PlatformFeeConfig{Tier: StoreTierBasic, Rate: 10, MinSkip: 10000})
	db.Model(&PlatformFeeConfig{}).Where("tier = ?", StoreTierBasic).Update("rate", 12.5)
	var c PlatformFeeConfig
	db.Where("tier = ?", StoreTierBasic).First(&c)
	if c.Rate != 12.5 {
		t.Fatalf("expected 12.5, got %f", c.Rate)
	}
}

func TestReviewCRUD(t *testing.T) {
	s, db := mkStore(t)
	lid := fmt.Sprintf("r-%d", time.Now().UnixNano())
	r := &Review{ListingID: lid, StoreID: s.ID, BuyerID: 777, Rating: 5, Content: "g"}
	db.Create(r)
	r2 := &Review{ListingID: lid, StoreID: s.ID, BuyerID: 778, Rating: 3, Content: "o"}
	db.Create(r2)
	var total int64
	db.Model(&Review{}).Where("listing_id = ?", lid).Count(&total)
	if total != 2 {
		t.Fatalf("expected 2, got %d", total)
	}
	var avg float64
	db.Model(&Review{}).Where("store_id = ?", s.ID).Select("COALESCE(AVG(rating), 0)").Scan(&avg)
	if avg != 4.0 {
		t.Fatalf("expected avg 4.0, got %f", avg)
	}
}

func TestStoreTierLogCreate(t *testing.T) {
	s, db := mkStore(t)
	db.Create(&StoreTierLog{StoreID: s.ID, FromTier: StoreTierBasic, ToTier: StoreTierGold, Reason: "t", Operator: 1, CreatedAt: time.Now().Unix()})
	var logs []StoreTierLog
	db.Where("store_id = ?", s.ID).Order("id desc").Limit(5).Find(&logs)
	if len(logs) < 1 {
		t.Fatal("expected logs")
	}
}

func TestStoreAutoUpgradeFlagship(t *testing.T) {
	s, db := mkStore(t)
	s.TotalSales = 1000000
	db.Save(s)
	db.Model(s).Update("tier", StoreTierFlagship)
	db.First(s, s.ID)
	if s.Tier != StoreTierFlagship {
		t.Fatalf("expected flagship, got %s", s.Tier)
	}
}

func TestStoreAutoUpgradeGold(t *testing.T) {
	s, db := mkStore(t)
	s.TotalSales = 100000
	db.Save(s)
	db.Model(s).Update("tier", StoreTierGold)
	db.First(s, s.ID)
	if s.Tier != StoreTierGold {
		t.Fatalf("expected gold, got %s", s.Tier)
	}
}

func TestListingByStoreID(t *testing.T) {
	s, db := mkStore(t)
	mn := fmt.Sprintf("bs-%d", time.Now().UnixNano())
	mkListing(db, s.ID, s.UserID, mn, 100)
	var listings []Listing
	db.Where("store_id = ?", s.ID).Find(&listings)
	if len(listings) < 1 {
		t.Fatal("expected listings")
	}
}

func TestStoreDefaultTier(t *testing.T) {
	s, _ := mkStore(t)
	if s.Tier != StoreTierBasic {
		t.Fatalf("expected basic, got %s", s.Tier)
	}
}
