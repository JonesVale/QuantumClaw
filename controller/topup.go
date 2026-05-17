package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/quantumclaw/quantumclaw/common"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/model"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// ==================== 安全增强的支付控制器 ====================

// TopUpRequest 充值请求（增强验证）
type TopUpRequest struct {
	Amount        int64  `json:"amount" binding:"required,min=1"`
	PaymentMethod string `json:"payment_method" binding:"required"`
	SuccessURL    string `json:"success_url,omitempty"`
	CancelURL     string `json:"cancel_url,omitempty"`
}

// TopUpResponse 充值响应
type TopUpResponse struct {
	TradeNo   string  `json:"trade_no"`
	Amount     int64   `json:"amount"`
	Money      float64 `json:"money"`
	PayURL     string  `json:"pay_url,omitempty"`
	CheckoutURL string  `json:"checkout_url,omitempty"`
	PayParams  string  `json:"pay_params,omitempty"`
}

// @Summary 获取支付信息
// @Description 获取可用的支付方式和配置
// @Tags User
// @Produce json
// @Success 200 {object} common.Response
// @Router /api/user/topup/info [get]
func GetTopUpInfo(c *gin.Context) {
	// 获取支付方式
	payMethods := common.GetPayMethods()

	data := gin.H{
		"enable_online_topup": common.IsEpayEnabled(),
		"enable_stripe_topup":  common.IsStripeEnabled(),
		"enable_creem_topup":  common.IsCreemEnabled(),
		"enable_waffo_topup":  common.IsWaffoEnabled(),
		"enable_binance_topup": common.IsBinanceEnabled(),
		"pay_methods":          payMethods,
		"min_topup":            common.GetMinTopUp(),
		"amount_options":       common.GetAmountOptions(),
		"discount":             common.GetAmountDiscount(),
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// @Summary 请求充值（Epay易支付）
// @Description 创建易支付订单
// @Tags User
// @Accept json
// @Produce json
// @Param request body TopUpRequest true "充值请求"
// @Success 200 {object} common.Response
// @Router /api/user/topup/epay [post]
func RequestEpayTopUp(c *gin.Context) {
	var req TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "参数错误", "data": err.Error()})
		return
	}

	// 验证支付方式
	if !common.IsValidPayMethod(req.PaymentMethod) {
		c.JSON(http.StatusOK, gin.H{"message": "不支持的支付方式"})
		return
	}

	// 验证金额
	minTopUp := common.GetMinTopUp()
	if req.Amount < int64(minTopUp) {
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("充值数量不能小于 %d", minTopUp)})
		return
	}

	// 获取用户信息
	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "获取用户信息失败"})
		return
	}

	// 计算支付金额（服务器端计算，防止客户端篡改）
	payMoney := calculatePayMoney(req.Amount, user.Group)

	// 创建订单
	tradeNo, err := model.GenerateSecureTradeNo(userId)
	if err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("生成订单号失败 user_id=%d error=%q", userId, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "创建订单失败"})
		return
	}

	topUp := &model.TopUp{
		UserId:          userId,
		Amount:           req.Amount,
		Money:            payMoney,
		TradeNo:          tradeNo,
		PaymentMethod:    req.PaymentMethod,
		PaymentProvider:  model.PaymentProviderEpay,
		UserIP:           c.ClientIP(),
		UserAgent:        c.GetHeader("User-Agent"),
	}

	if err := topUp.Insert(); err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("创建充值订单失败 user_id=%d trade_no=%s error=%q", userId, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "创建订单失败"})
		return
	}

	// TODO: 调用易支付SDK创建支付链接
	// 这里需要集成 go-epay 库

	logger.Info(c.Request.Context(), fmt.Sprintf("易支付订单创建成功 user_id=%d trade_no=%s amount=%d money=%.2f", 
		userId, tradeNo, req.Amount, payMoney))

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"trade_no": tradeNo,
			"amount":   req.Amount,
			"money":    payMoney,
		},
	})
}

// @Summary Epay支付回调
// @Description 处理易支付回调通知
// @Tags Payment
// @Accept json
// @Produce json
// @Success 200 {string} string "success"
// @Router /api/user/topup/epay/notify [post]
func EpayNotify(c *gin.Context) {
	// TODO: 实现Epay回调处理
	// 1. 验证签名
	// 2. 更新订单状态
	// 3. 充值用户配额
	logger.Info(c.Request.Context(), "Epay回调收到")
	c.String(http.StatusOK, "success")
}

// @Summary 查询充值订单
// @Description 查询用户的充值订单列表
// @Tags User
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} common.Response
// @Router /api/user/topup/list [get]
func GetTopUpList(c *gin.Context) {
	userId := c.GetInt("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	topUps, total, err := model.GetUserTopUps(userId, page, pageSize)
	if err != nil {
		logger.Error(c.Request.Context(), fmt.Sprintf("查询充值订单失败 user_id=%d error=%q", userId, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"items":      topUps,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
		},
	})
}

// ==================== 辅助函数（安全增强） ====================

// calculatePayMoney 计算支付金额（服务器端计算，防止篡改）
func calculatePayMoney(amount int64, group string) float64 {
	dAmount := decimal.NewFromInt(amount)
	
	// 获取价格配置
	price := decimal.NewFromFloat(common.GetPrice())
	
	// 获取用户组比例
	groupRatio := common.GetTopupGroupRatio(group)
	if groupRatio == 0 {
		groupRatio = 1
	}
	dGroupRatio := decimal.NewFromFloat(groupRatio)
	
	// 计算金额
	payMoney := dAmount.Mul(price).Mul(dGroupRatio)
	
	// 应用折扣（如果有）
	discount := getDiscount(amount)
	dDiscount := decimal.NewFromFloat(discount)
	
	result := payMoney.Mul(dDiscount)
	return result.InexactFloat64()
}

// calculateStripePayMoney 计算Stripe支付金额
func calculateStripePayMoney(amount int64, group string) float64 {
	// 类似calculatePayMoney，但使用Stripe的价格配置
	return calculatePayMoney(amount, group)
}

// getDiscount 获取折扣（从配置中读取）
func getDiscount(amount int64) float64 {
	discountMap := common.GetAmountDiscount()
	if discount, ok := discountMap[int(amount)]; ok && discount > 0 {
		return discount
	}
	return 1.0
}

// validatePaymentAmount 验证支付金额（防止负数和异常值）
func validatePaymentAmount(amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("充值金额必须大于0")
	}
	
	if amount > 1000000 {
		return fmt.Errorf("充值金额过大")
	}
	
	return nil
}

// rateLimitKey 生成限流key（防止暴力请求）
func rateLimitKey(userId int, paymentMethod string) string {
	return fmt.Sprintf("topup:%d:%s", userId, paymentMethod)
}
