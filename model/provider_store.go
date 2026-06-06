package model

import (
	"time"

	"github.com/quantumclaw/quantumclaw/common/logger"
)

// ProviderStore 渠道商店铺
type ProviderStore struct {
	Id          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId      int       `json:"user_id" gorm:"uniqueIndex;not null"`               // 渠道商用户ID
	Name        string    `json:"name" gorm:"type:varchar(100)"`                     // 店铺名称
	Description string    `json:"description" gorm:"type:text"`                      // 店铺描述
	Logo        string    `json:"logo" gorm:"type:varchar(500)"`                     // Logo URL
	BannerUrl   string    `json:"banner_url" gorm:"type:varchar(500)"`               // Banner 图片
	ContactInfo string    `json:"contact_info" gorm:"type:varchar(500)"`             // 联系方式
	StoreSlug   string    `json:"store_slug" gorm:"type:varchar(50);uniqueIndex"`    // 唯一路径 /store/{slug}
	Status      int       `json:"status" gorm:"default:1"`                           // 1=营业 0=歇业
	Rating      float64   `json:"rating" gorm:"type:decimal(3,2);default:0"`         // 评分
	TotalSales  int64     `json:"total_sales" gorm:"bigint;default:0"`               // 累计成交数
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// StoreModel 店铺上架模型
type StoreModel struct {
	Id              int       `json:"id" gorm:"primaryKey;autoIncrement"`
	StoreId         int       `json:"store_id" gorm:"index;not null"`                // 关联店铺
	ChannelId       int       `json:"channel_id" gorm:"not null"`                    // 关联的 API 渠道
	ModelName       string    `json:"model_name" gorm:"type:varchar(100);not null"`  // 模型标识
	DisplayName     string    `json:"display_name" gorm:"type:varchar(100)"`         // 展示名称
	Description     string    `json:"description" gorm:"type:text"`                  // 模型说明
	Tags            string    `json:"tags" gorm:"type:varchar(200)"`                 // 逗号分隔标签（如"代码编程,多模态,推理"）
	InputPrice      int64     `json:"input_price" gorm:"bigint;default:0"`           // 自定义输入价格(分/1M tokens, 0=使用倍率)
	OutputPrice     int64     `json:"output_price" gorm:"bigint;default:0"`          // 自定义输出价格
	CacheReadPrice  int64     `json:"cache_read_price" gorm:"bigint;default:0"`      // 缓存读取价格
	PriceMultiplier float64   `json:"price_multiplier" gorm:"type:decimal(5,3);default:1.0"` // 基础价倍率
	IsActive        bool      `json:"is_active" gorm:"default:true"`                 // true=上架 false=下架
	SortOrder       int       `json:"sort_order" gorm:"default:0"`                   // 排序权重
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// StoreWithOwner 店铺+店主信息（用于公开市场展示）
type StoreWithOwner struct {
	ProviderStore
	Username    string `json:"username"`
	ModelCount  int64  `json:"model_count"`
	ActiveModel int64  `json:"active_model"`
}

func InitStoreTables() {
	if err := DB.AutoMigrate(&ProviderStore{}); err != nil {
		logger.SysError("InitStoreTables ProviderStore AutoMigrate failed: " + err.Error())
		return
	}
	if err := DB.AutoMigrate(&StoreModel{}); err != nil {
		logger.SysError("InitStoreTables StoreModel AutoMigrate failed: " + err.Error())
		return
	}
	logger.SysLog("store tables initialized")
}

// ── 店铺 CRUD ──

// GetOrCreateStore 获取或创建渠道商的店铺（懒创建）
func GetOrCreateStore(userId int) (*ProviderStore, error) {
	var store ProviderStore
	if err := DB.Where("user_id = ?", userId).First(&store).Error; err != nil {
		// 自动创建默认店铺
		store = ProviderStore{
			UserId:    userId,
			Name:      "",
			StoreSlug: "",
			Status:    1,
		}
		if err := DB.Create(&store).Error; err != nil {
			return nil, err
		}
	}
	return &store, nil
}

// GetStoreBySlug 根据路径获取店铺
func GetStoreBySlug(slug string) (*ProviderStore, error) {
	var store ProviderStore
	if err := DB.Where("store_slug = ?", slug).First(&store).Error; err != nil {
		return nil, err
	}
	return &store, nil
}

// UpdateStore 更新店铺信息
func UpdateStore(store *ProviderStore) error {
	return DB.Model(store).Select("name", "description", "logo", "banner_url", "contact_info", "store_slug", "status").Updates(store).Error
}

// GetAllActiveStores 获取所有营业中店铺（含模型数）
func GetAllActiveStores() ([]StoreWithOwner, error) {
	var result []StoreWithOwner
	err := DB.Raw(`
		SELECT s.*, u.username,
			(SELECT COUNT(*) FROM store_models sm WHERE sm.store_id = s.id) AS model_count,
			(SELECT COUNT(*) FROM store_models sm WHERE sm.store_id = s.id AND sm.is_active = 1) AS active_model
		FROM provider_stores s
		JOIN users u ON u.id = s.user_id
		WHERE s.status = 1
		ORDER BY s.rating DESC, s.total_sales DESC
	`).Scan(&result).Error
	return result, err
}

// ── 商品管理 ──

// GetStoreModels 获取店铺上架的模型列表
func GetStoreModels(storeId int, activeOnly bool) ([]StoreModel, error) {
	query := DB.Where("store_id = ?", storeId)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	var models []StoreModel
	err := query.Order("sort_order ASC, id ASC").Find(&models).Error
	return models, err
}

// ListStoreModelsByTag 按标签筛选
func ListStoreModelsByTag(storeId int, tag string) ([]StoreModel, error) {
	var models []StoreModel
	err := DB.Where("store_id = ? AND is_active = ? AND tags LIKE ?", storeId, true, "%"+tag+"%").
		Order("sort_order ASC").Find(&models).Error
	return models, err
}

// AddStoreModel 上架模型
func AddStoreModel(sm *StoreModel) error {
	return DB.Create(sm).Error
}

// UpdateStoreModel 更新模型信息
func UpdateStoreModel(sm *StoreModel) error {
	return DB.Model(sm).Select("display_name", "description", "tags", "input_price", "output_price", "cache_read_price", "price_multiplier", "is_active", "sort_order").Updates(sm).Error
}

// DeleteStoreModel 下架（物理删除）
func DeleteStoreModel(id int) error {
	return DB.Delete(&StoreModel{}, id).Error
}

// ToggleStoreModel 切换上架/下架
func ToggleStoreModel(id int) error {
	return DB.Model(&StoreModel{}).Where("id = ?", id).
		Update("is_active", DB.Raw("NOT is_active")).Error
}

// CountStoreModels 统计店铺模型数
func CountStoreModels(storeId int) (int64, error) {
	var count int64
	err := DB.Model(&StoreModel{}).Where("store_id = ?", storeId).Count(&count).Error
	return count, err
}
