package model

import "gorm.io/gorm"

// InitMarketTables initializes marketplace-related database tables.
// Kept separate from the main init to avoid encoding conflicts in model/main.go.
func InitMarketTables(gdb *gorm.DB) {
	gdb.AutoMigrate(&Store{}, &Listing{}, &StoreTierLog{}, &PlatformFeeConfig{}, &Review{})
	gdb.AutoMigrate(&ConsumeRecord{}, &PaymentIdempotencyKey{})
	gdb.AutoMigrate(&PlatformPoolAgreement{}, &UserPoolConsent{})
}

// SeedMarketDefaults initializes default configs for marketplace features.
func SeedMarketDefaults() {
	SeedDefaultFeeConfigs()
	SeedDefaultAgreement()
}
