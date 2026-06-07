package model

import (
	"time"

	"github.com/quantumclaw/quantumclaw/common/helper"
	"gorm.io/gorm"
)

type StoreTier string
const (
	StoreTierBasic    StoreTier = "basic"
	StoreTierGold     StoreTier = "gold"
	StoreTierFlagship StoreTier = "flagship"
)

type StoreStatus string
const (
	StoreStatusActive    StoreStatus = "active"
	StoreStatusSuspended StoreStatus = "suspended"
	StoreStatusClosed    StoreStatus = "closed"
)

type ListingStatus string
const (
	ListingStatusActive   ListingStatus = "active"
	ListingStatusPaused   ListingStatus = "paused"
	ListingStatusArchived ListingStatus = "archived"
)

type Store struct {
	ID           int         `json:"id" gorm:"primaryKey"`
	UserID       int         `json:"user_id" gorm:"uniqueIndex;not null"`
	Name         string      `json:"name" gorm:"type:varchar(128)"`
	Tier         StoreTier   `json:"tier" gorm:"type:varchar(32);default:'basic'"`
	Status       StoreStatus `json:"status" gorm:"type:varchar(32);default:'active'"`
	Rating       float64     `json:"rating" gorm:"type:decimal(3,2);default:5.0"`
	TotalSales   int64       `json:"total_sales" gorm:"bigint;default:0"`
	TotalOrders  int64       `json:"total_orders" gorm:"bigint;default:0"`
	OpenedAt     int64       `json:"opened_at" gorm:"bigint"`
	LastActiveAt int64       `json:"last_active_at" gorm:"bigint"`
}

func (Store) TableName() string { return "stores" }

type Listing struct {
	ID            string        `json:"id" gorm:"primaryKey;type:varchar(64)"`
	StoreID       int           `json:"store_id" gorm:"index;not null"`
	ProviderID    int           `json:"provider_id" gorm:"index;not null"`
	ChannelID     int           `json:"channel_id" gorm:"index;not null"`
	ModelName     string        `json:"model_name" gorm:"type:varchar(128);index;not null"`
	Region        string        `json:"region" gorm:"type:varchar(32);default:'overseas'"`
	PricePerUnit  int64         `json:"price_per_unit" gorm:"bigint;not null"`
	Unit          string        `json:"unit" gorm:"type:varchar(32);default:'1k_tokens'"`
	MinPurchase   int64         `json:"min_purchase" gorm:"bigint;default:1"`
	MaxConcurrent int           `json:"max_concurrent" gorm:"int;default:10"`
	Status        ListingStatus `json:"status" gorm:"type:varchar(32);default:'active'"`
	AvgLatencyMs  float64       `json:"avg_latency_ms" gorm:"type:decimal(10,2);default:0"`
	Availability  float64       `json:"availability" gorm:"type:decimal(5,2);default:100.00"`
	TotalOrders   int64         `json:"total_orders" gorm:"bigint;default:0"`
	AvgRating     float64       `json:"avg_rating" gorm:"type:decimal(3,2);default:5.0"`
	CreatedAt     int64         `json:"created_at" gorm:"bigint"`
	UpdatedAt     int64         `json:"updated_at" gorm:"bigint"`
}

func (Listing) TableName() string { return "listings" }

type StoreTierLog struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	StoreID   int       `json:"store_id" gorm:"index;not null"`
	FromTier  StoreTier `json:"from_tier" gorm:"type:varchar(32)"`
	ToTier    StoreTier `json:"to_tier" gorm:"type:varchar(32)"`
	Reason    string    `json:"reason" gorm:"type:varchar(64)"`
	Operator  int       `json:"operator" gorm:"int;default:0"`
	CreatedAt int64     `json:"created_at" gorm:"bigint"`
}

func (StoreTierLog) TableName() string { return "store_tier_logs" }

type Review struct {
	ID        int    `json:"id" gorm:"primaryKey"`
	ListingID string `json:"listing_id" gorm:"index;not null"`
	StoreID   int    `json:"store_id" gorm:"index;not null"`
	BuyerID   int    `json:"buyer_id" gorm:"index;not null"`
	OrderID   string `json:"order_id" gorm:"type:varchar(64)"`
	Rating    int    `json:"rating" gorm:"type:tinyint;not null"`
	Content   string `json:"content" gorm:"type:text"`
	CreatedAt int64  `json:"created_at" gorm:"bigint"`
}

func (Review) TableName() string { return "reviews" }

// ---- Store CRUD ----

func CreateStore(userID int, name string) (*Store, error) {
	store := &Store{
		UserID:       userID,
		Name:         name,
		Tier:         StoreTierBasic,
		Status:       StoreStatusActive,
		Rating:       5.0,
		OpenedAt:     helper.GetTimestamp(),
		LastActiveAt: helper.GetTimestamp(),
	}
	if err := DB.Create(store).Error; err != nil {
		return nil, err
	}
	CreateStoreTierLog(store.ID, "", StoreTierBasic, "store_opened", 0)
	return store, nil
}

func GetStoreByUserID(userID int) (*Store, error) {
	var store Store
	err := DB.Where("user_id = ?", userID).First(&store).Error
	return &store, err
}

func GetStoreByID(id int) (*Store, error) {
	var store Store
	err := DB.Where("id = ?", id).First(&store).Error
	return &store, err
}

func UpdateStoreInfo(store *Store) error {
	return DB.Model(store).Updates(map[string]interface{}{
		"name":          store.Name,
		"tier":          store.Tier,
		"status":        store.Status,
		"rating":        store.Rating,
		"total_sales":   store.TotalSales,
		"total_orders":  store.TotalOrders,
		"last_active_at": helper.GetTimestamp(),
	}).Error
}

func GetActiveStores() ([]Store, error) {
	var stores []Store
	err := DB.Where("status = ?", StoreStatusActive).Find(&stores).Error
	return stores, err
}

func GetAllStores(page, pageSize int) ([]Store, int64, error) {
	var stores []Store
	var total int64
	DB.Model(&Store{}).Count(&total)
	err := DB.Order("id desc").Offset((page-1)*pageSize).Limit(pageSize).Find(&stores).Error
	return stores, total, err
}

// ---- Listing CRUD ----

func CreateListing(listing *Listing) error {
	return DB.Create(listing).Error
}

func GetListingByID(id string) (*Listing, error) {
	var listing Listing
	err := DB.Where("id = ?", id).First(&listing).Error
	return &listing, err
}

func GetListingsByStoreID(storeID int) ([]Listing, error) {
	var listings []Listing
	err := DB.Where("store_id = ?", storeID).Order("created_at desc").Find(&listings).Error
	return listings, err
}

func UpdateListing(listing *Listing) error {
	return DB.Model(listing).Updates(map[string]interface{}{
		"price_per_unit": listing.PricePerUnit,
		"status":         listing.Status,
		"min_purchase":   listing.MinPurchase,
		"max_concurrent": listing.MaxConcurrent,
		"avg_latency_ms": listing.AvgLatencyMs,
		"availability":   listing.Availability,
		"total_orders":   listing.TotalOrders,
		"avg_rating":     listing.AvgRating,
		"updated_at":     helper.GetTimestamp(),
	}).Error
}

func SearchActiveListings(modelName, region string, limit int) ([]Listing, error) {
	var listings []Listing
	query := DB.Where("status = ? AND model_name = ?", ListingStatusActive, modelName)
	if region != "" {
		query = query.Where("region = ?", region)
	}
	err := query.Order("price_per_unit ASC").Limit(limit).Find(&listings).Error
	return listings, err
}

func GetCheapestListing(modelName string) (*Listing, error) {
	var listing Listing
	err := DB.Where("status = ? AND model_name = ?", ListingStatusActive, modelName).
		Order("price_per_unit ASC").Limit(1).First(&listing).Error
	return &listing, err
}

func GetAllActiveListingsForModel(modelName string) ([]Listing, error) {
	var listings []Listing
	err := DB.Where("status = ? AND model_name = ?", ListingStatusActive, modelName).
		Order("price_per_unit ASC").Find(&listings).Error
	return listings, err
}

func GetAllActiveListings() ([]Listing, error) {
	var listings []Listing
	err := DB.Where("status = ?", ListingStatusActive).Find(&listings).Error
	return listings, err
}

// ---- StoreTierLog CRUD ----

func CreateStoreTierLog(storeID int, fromTier, toTier StoreTier, reason string, operator int) error {
	return DB.Create(&StoreTierLog{
		StoreID:   storeID,
		FromTier:  fromTier,
		ToTier:    toTier,
		Reason:    reason,
		Operator:  operator,
		CreatedAt: helper.GetTimestamp(),
	}).Error
}

func GetStoreTierLogs(storeID int, limit int) ([]StoreTierLog, error) {
	var logs []StoreTierLog
	err := DB.Where("store_id = ?", storeID).Order("id desc").Limit(limit).Find(&logs).Error
	return logs, err
}

// ---- Review CRUD ----

func CreateReview(review *Review) error {
	return DB.Create(review).Error
}

func GetReviewsByListingID(listingID string, limit, offset int) ([]Review, int64, error) {
	var reviews []Review
	var total int64
	DB.Model(&Review{}).Where("listing_id = ?", listingID).Count(&total)
	err := DB.Where("listing_id = ?", listingID).Order("id desc").Limit(limit).Offset(offset).Find(&reviews).Error
	return reviews, total, err
}

func GetAverageRatingByStoreID(storeID int) (float64, error) {
	var avg float64
	err := DB.Model(&Review{}).Where("store_id = ?", storeID).
		Select("COALESCE(AVG(rating), 5.0)").Scan(&avg).Error
	return avg, err
}

// ---- Business Logic ----

func CalculateStoreFeeRate(storeID int) float64 {
	var store Store
	if err := DB.First(&store, storeID).Error; err != nil {
		return 10.0
	}
	var cfg PlatformFeeConfig
	if err := DB.Where("tier = ?", store.Tier).First(&cfg).Error; err != nil {
		return 10.0
	}
	return cfg.Rate
}

func AutoUpgradeStoreTier(storeID int, totalSales int64) (bool, error) {
	var store Store
	if err := DB.First(&store, storeID).Error; err != nil {
		return false, err
	}

	newTier := store.Tier
	switch {
	case totalSales >= 1000000:
		newTier = StoreTierFlagship
	case totalSales >= 100000:
		newTier = StoreTierGold
	case store.LastActiveAt < time.Now().AddDate(0, -3, 0).Unix():
		newTier = StoreTierBasic
	}

	if newTier != store.Tier {
		oldTier := store.Tier
		store.Tier = newTier
		if err := DB.Model(&store).Update("tier", newTier).Error; err != nil {
			return false, err
		}
		reason := "auto_upgrade"
		if newTier == StoreTierBasic && store.LastActiveAt < time.Now().AddDate(0, -3, 0).Unix() {
			reason = "auto_downgrade"
		}
		CreateStoreTierLog(storeID, oldTier, newTier, reason, 0)
		return true, nil
	}
	return false, nil
}

func UpdateStoreStats(storeID int, orderAmount int64) error {
	return DB.Model(&Store{}).Where("id = ?", storeID).Updates(map[string]interface{}{
		"total_sales":    gorm.Expr("total_sales + ?", orderAmount),
		"total_orders":   gorm.Expr("total_orders + 1"),
		"last_active_at": helper.GetTimestamp(),
	}).Error
}

func UpdateListingStats(listingID string, latencyMs float64, orderAmount int64) error {
	return DB.Model(&Listing{}).Where("id = ?", listingID).Updates(map[string]interface{}{
		"total_orders":   gorm.Expr("total_orders + 1"),
		"avg_latency_ms": gorm.Expr("(avg_latency_ms * total_orders + ?) / (total_orders + 1)", latencyMs),
	}).Error
}
