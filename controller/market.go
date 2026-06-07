package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"
)

// ?? ???? ??

// GetMarketModels ?????????? + ???
func GetMarketModels(c *gin.Context) {
	// ???????? API
	models, err := model.GetAllActiveListings()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	// ??????????
	type ModelPrice struct {
		ModelName    string `json:"model_name"`
		MinPrice     int64  `json:"min_price"`
		ProviderCount int   `json:"provider_count"`
	}
	modelMap := make(map[string]*ModelPrice)
	var modelOrder []string
	for _, m := range models {
		if _, ok := modelMap[m.ModelName]; !ok {
			modelMap[m.ModelName] = &ModelPrice{
				ModelName: m.ModelName,
				MinPrice:  m.PricePerUnit,
			}
			modelOrder = append(modelOrder, m.ModelName)
		}
		mp := modelMap[m.ModelName]
		if m.PricePerUnit < mp.MinPrice {
			mp.MinPrice = m.PricePerUnit
		}
		mp.ProviderCount++
	}
	var result []*ModelPrice
	for _, name := range modelOrder {
		result = append(result, modelMap[name])
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetMarketPrice ????????????
func GetMarketPrice(c *gin.Context) {
	modelName := c.Param("model")
	listings, err := model.GetAllActiveListingsForModel(modelName)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	type StoreQuote struct {
		ListingID    string  `json:"listing_id"`
		StoreID      int     `json:"store_id"`
		PricePerUnit int64   `json:"price_per_unit"`
		Region       string  `json:"region"`
		AvgRating    float64 `json:"avg_rating"`
		AvgLatencyMs float64 `json:"avg_latency_ms"`
		Availability float64 `json:"availability"`
		TotalOrders  int64   `json:"total_orders"`
	}
	var quotes []StoreQuote
	for _, l := range listings {
		store, _ := model.GetStoreByID(l.StoreID)
		rating := l.AvgRating
		if store != nil {
			rating = store.Rating
		}
		quotes = append(quotes, StoreQuote{
			ListingID:    l.ID,
			StoreID:      l.StoreID,
			PricePerUnit: l.PricePerUnit,
			Region:       l.Region,
			AvgRating:    rating,
			AvgLatencyMs: l.AvgLatencyMs,
			Availability: l.Availability,
			TotalOrders:  l.TotalOrders,
		})
	}
	if quotes == nil {
		quotes = []StoreQuote{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"model":  modelName,
			"quotes": quotes,
		},
	})
}

// GetMarketStores ???????????
func GetMarketStores(c *gin.Context) {
	modelName := c.Query("model")
	region := c.Query("region")
	sortBy := c.DefaultQuery("sort", "price") // price / rating

	listings, err := model.SearchActiveListings(modelName, region, 50)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "data": []model.Listing{}})
		return
	}

	// ? sortBy ??
	switch sortBy {
	case "rating":
		for i := 0; i < len(listings); i++ {
			for j := i + 1; j < len(listings); j++ {
				if listings[j].AvgRating > listings[i].AvgRating {
					listings[i], listings[j] = listings[j], listings[i]
				}
			}
		}
	default: // price
		for i := 0; i < len(listings); i++ {
			for j := i + 1; j < len(listings); j++ {
				if listings[j].PricePerUnit < listings[i].PricePerUnit {
					listings[i], listings[j] = listings[j], listings[i]
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": listings})
}

// GetMarketListing ??????
func GetMarketListing(c *gin.Context) {
	id := c.Param("id")
	listing, err := model.GetListingByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "???"})
		return
	}
	store, _ := model.GetStoreByID(listing.StoreID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"listing": listing,
			"store":   store,
		},
	})
}

// ?? ?? ??

// CreateReview ????
func CreateReview(c *gin.Context) {
	userID := c.GetInt("id")
	var req struct {
		ListingID string `json:"listing_id" binding:"required"`
		OrderID   string `json:"order_id"`
		Rating    int    `json:"rating" binding:"required,min=1,max=5"`
		Content   string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "????"})
		return
	}

	listing, err := model.GetListingByID(req.ListingID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "?????"})
		return
	}

	review := &model.Review{
		ListingID: req.ListingID,
		StoreID:   listing.StoreID,
		BuyerID:   userID,
		OrderID:   req.OrderID,
		Rating:    req.Rating,
		Content:   req.Content,
	}
	if err := model.CreateReview(review); err != nil {
		logger.Errorf(c.Request.Context(), "create review failed: %v", err)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "????"})
		return
	}

	// ??????
	avgRating, _ := model.GetAverageRatingByStoreID(listing.StoreID)
	model.DB.Model(&model.Store{}).Where("id = ?", listing.StoreID).Update("rating", avgRating)
	// ??????
	model.DB.Model(&model.Listing{}).Where("id = ?", req.ListingID).Update("avg_rating", avgRating)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "?????"})
}

// GetReviews ????
func GetReviews(c *gin.Context) {
	listingID := c.Query("listing_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize > 50 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	reviews, total, err := model.GetReviewsByListingID(listingID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "data": []model.Review{}})
		return
	}
	if reviews == nil {
		reviews = []model.Review{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reviews,
		"total":   total,
		"page":    page,
	})
}

// ?? ???? ??

// SetPreferredStore ???????
func SetPreferredStore(c *gin.Context) {
	userID := c.GetInt("id")
	var req struct {
		StoreID int `json:"store_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "????"})
		return
	}

	// ?????????
	store, err := model.GetStoreByID(req.StoreID)
	if err != nil || store.Status != model.StoreStatusActive {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "?????????"})
		return
	}

	// ???? preferences??? option ??? user ???????
	model.UpdateOption("preferred_store_"+strconv.Itoa(userID), strconv.Itoa(req.StoreID))

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "???"})
}

// GetPreferredStore ???????
func GetPreferredStore(c *gin.Context) {
	userID := c.GetInt("id")
	var val string
	model.DB.Model(&model.Option{}).Select("value").Where("`key` = ?", "preferred_store_"+strconv.Itoa(userID)).Scan(&val)
	if val == "" {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
		return
	}
	storeID, _ := strconv.Atoi(val)
	store, err := model.GetStoreByID(storeID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": store})
}
