package router

import (
	"github.com/quantumclaw/quantumclaw/controller"
	"github.com/quantumclaw/quantumclaw/controller/auth"
	"github.com/quantumclaw/quantumclaw/middleware"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetApiRouter(router *gin.Engine) {
	apiRouter := router.Group("/api")
	apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))
	apiRouter.Use(middleware.GlobalAPIRateLimit())
	{
		apiRouter.GET("/status", controller.GetStatus)
		apiRouter.GET("/models", middleware.UserAuth(), controller.DashboardListModels)
		apiRouter.GET("/notice", controller.GetNotice)
		apiRouter.GET("/about", controller.GetAbout)
		apiRouter.GET("/home_page_content", controller.GetHomePageContent)

		// 支付 Webhook 回调路由（不需要用户认证）
		apiRouter.Any("/webhook/epay", controller.EpayNotify)             // 易支付回调
		apiRouter.POST("/webhook/stripe", controller.StripeWebhook)       // Stripe Webhook
		apiRouter.POST("/webhook/creem", controller.CreemWebhook)       // Creem Webhook
		apiRouter.POST("/webhook/waffo", controller.WaffoWebhook)       // Waffo Webhook
		apiRouter.GET("/verification", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendEmailVerification)
		apiRouter.GET("/reset_password", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendPasswordResetEmail)
		apiRouter.POST("/user/reset", middleware.CriticalRateLimit(), controller.ResetPassword)
		apiRouter.GET("/oauth/github", middleware.CriticalRateLimit(), auth.GitHubOAuth)
		apiRouter.GET("/oauth/oidc", middleware.CriticalRateLimit(), auth.OidcAuth)
		apiRouter.GET("/oauth/lark", middleware.CriticalRateLimit(), auth.LarkOAuth)
		apiRouter.GET("/oauth/state", middleware.CriticalRateLimit(), auth.GenerateOAuthCode)
		apiRouter.GET("/oauth/wechat", middleware.CriticalRateLimit(), auth.WeChatAuth)
		apiRouter.GET("/oauth/wechat/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), auth.WeChatBind)
		apiRouter.GET("/oauth/email/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), controller.EmailBind)
		apiRouter.POST("/topup", middleware.AdminAuth(), controller.AdminTopUp)

		userRoute := apiRouter.Group("/user")
		{
			userRoute.POST("/register", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.Register)
			userRoute.POST("/login", middleware.CriticalRateLimit(), controller.Login)
			userRoute.GET("/logout", controller.Logout)

			selfRoute := userRoute.Group("/")
			selfRoute.Use(middleware.UserAuth())
			{
				selfRoute.GET("/dashboard", controller.GetUserDashboard)
				selfRoute.GET("/self", controller.GetSelf)
				selfRoute.PUT("/self", controller.UpdateSelf)
				selfRoute.DELETE("/self", controller.DeleteSelf)
				selfRoute.GET("/token", controller.GenerateAccessToken)
				selfRoute.GET("/aff", controller.GetAffCode)
				selfRoute.POST("/topup", controller.TopUp)
				selfRoute.GET("/available_models", controller.GetUserAvailableModels)
				
				// 支付相关路由（安全增强版）
				selfRoute.GET("/topup/info", controller.GetTopUpInfo)                              // 获取支付信息
				selfRoute.POST("/topup/epay", controller.RequestEpayTopUp)                      // 易支付
				selfRoute.POST("/topup/stripe", controller.RequestStripeTopUp)                  // Stripe支付
				selfRoute.POST("/topup/creem", controller.RequestCreemTopUp)                   // Creem支付
				selfRoute.POST("/topup/waffo", controller.RequestWaffoTopUp)                   // Waffo支付
				selfRoute.GET("/topup/list", controller.GetTopUpList)                          // 查询订单列表

				// 签到相关路由
				selfRoute.GET("/checkin", controller.GetCheckinStatus)  // 获取签到状态
				selfRoute.POST("/checkin", controller.DoCheckin)        // 执行签到

				// 订阅相关路由（用户端）
				selfRoute.GET("/subscription/plans", controller.GetSubscriptionPlans) // 获取套餐列表
				selfRoute.GET("/subscription/self", controller.GetSubscriptionSelf)   // 获取个人订阅
			}

			adminRoute := userRoute.Group("/")
			adminRoute.Use(middleware.AdminAuth())
			{
				adminRoute.GET("/", controller.GetAllUsers)
				adminRoute.GET("/search", controller.SearchUsers)
				adminRoute.GET("/:id", controller.GetUser)
				adminRoute.POST("/", controller.CreateUser)
				adminRoute.POST("/manage", controller.ManageUser)
				adminRoute.PUT("/", controller.UpdateUser)
				adminRoute.DELETE("/:id", controller.DeleteUser)
			}
		}
		optionRoute := apiRouter.Group("/option")
		optionRoute.Use(middleware.RootAuth())
		{
			optionRoute.GET("/", controller.GetOptions)
			optionRoute.PUT("/", controller.UpdateOption)
		}
		channelRoute := apiRouter.Group("/channel")
		channelRoute.Use(middleware.AdminAuth())
		{
			channelRoute.GET("/", controller.GetAllChannels)
			channelRoute.GET("/search", controller.SearchChannels)
			channelRoute.GET("/models", controller.ListAllModels)
			channelRoute.GET("/:id", controller.GetChannel)
			channelRoute.GET("/test", controller.TestChannels)
			channelRoute.GET("/test/:id", controller.TestChannel)
			channelRoute.GET("/update_balance", controller.UpdateAllChannelsBalance)
			channelRoute.GET("/update_balance/:id", controller.UpdateChannelBalance)
			channelRoute.POST("/", controller.AddChannel)
			channelRoute.PUT("/", controller.UpdateChannel)
			channelRoute.DELETE("/disabled", controller.DeleteDisabledChannel)
			channelRoute.DELETE("/:id", controller.DeleteChannel)
		}
		tokenRoute := apiRouter.Group("/token")
		tokenRoute.Use(middleware.UserAuth())
		{
			tokenRoute.GET("/", controller.GetAllTokens)
			tokenRoute.GET("/search", controller.SearchTokens)
			tokenRoute.GET("/:id", controller.GetToken)
			tokenRoute.POST("/", controller.AddToken)
			tokenRoute.PUT("/", controller.UpdateToken)
			tokenRoute.DELETE("/:id", controller.DeleteToken)
		}
		redemptionRoute := apiRouter.Group("/redemption")
		redemptionRoute.Use(middleware.AdminAuth())
		{
			redemptionRoute.GET("/", controller.GetAllRedemptions)
			redemptionRoute.GET("/search", controller.SearchRedemptions)
			redemptionRoute.GET("/:id", controller.GetRedemption)
			redemptionRoute.POST("/", controller.AddRedemption)
			redemptionRoute.PUT("/", controller.UpdateRedemption)
			redemptionRoute.DELETE("/:id", controller.DeleteRedemption)
		}
		logRoute := apiRouter.Group("/log")
		logRoute.GET("/", middleware.AdminAuth(), controller.GetAllLogs)
		logRoute.DELETE("/", middleware.AdminAuth(), controller.DeleteHistoryLogs)
		logRoute.GET("/stat", middleware.AdminAuth(), controller.GetLogsStat)
		logRoute.GET("/self/stat", middleware.UserAuth(), controller.GetLogsSelfStat)
		logRoute.GET("/search", middleware.AdminAuth(), controller.SearchAllLogs)
		logRoute.GET("/self", middleware.UserAuth(), controller.GetUserLogs)
		logRoute.GET("/self/search", middleware.UserAuth(), controller.SearchUserLogs)
		groupRoute := apiRouter.Group("/group")
		groupRoute.Use(middleware.AdminAuth())
		{
			groupRoute.GET("/", controller.GetGroups)
		}

		// 订阅套餐管理（管理员）
		subscriptionRoute := apiRouter.Group("/subscription")
		{
			// 用户端（需登录）
			subscriptionRoute.Use(middleware.UserAuth())
			subscriptionRoute.GET("/plans", controller.GetSubscriptionPlans)
			subscriptionRoute.GET("/self", controller.GetSubscriptionSelf)
		}

		// 管理员订阅套餐 CRUD
		adminSubscriptionRoute := apiRouter.Group("/admin/subscription")
		adminSubscriptionRoute.Use(middleware.AdminAuth())
		{
			adminSubscriptionRoute.GET("/plans", controller.AdminListSubscriptionPlans)
			adminSubscriptionRoute.POST("/plans", controller.AdminCreateSubscriptionPlan)
			adminSubscriptionRoute.PUT("/plans/:id", controller.AdminUpdateSubscriptionPlan)
			adminSubscriptionRoute.PUT("/plans/:id/status", controller.AdminUpdateSubscriptionPlanStatus)
			adminSubscriptionRoute.DELETE("/plans/:id", controller.AdminDeleteSubscriptionPlan)
			adminSubscriptionRoute.POST("/bind", controller.AdminBindSubscription)
			adminSubscriptionRoute.GET("/user/:id", controller.AdminListUserSubscriptions)
			adminSubscriptionRoute.POST("/user/:id", controller.AdminCreateUserSubscription)
			adminSubscriptionRoute.PUT("/user-sub/:id/invalidate", controller.AdminInvalidateUserSubscription)
			adminSubscriptionRoute.DELETE("/user-sub/:id", controller.AdminDeleteUserSubscription)
		}

		// ==================== 支付 Webhook 回调 ====================
		apiRouter.POST("/webhook/waffo_pancake", controller.HandleWaffoPancakeWebhook)

		// ==================== 2FA 两步验证路由 ====================
		userTwoFARoute := apiRouter.Group("/user/2fa")
		userTwoFARoute.Use(middleware.UserAuth())
		{
			userTwoFARoute.GET("/", controller.GetTwoFAStatus)
			userTwoFARoute.POST("/init", controller.InitTwoFA)
			userTwoFARoute.POST("/enable", controller.VerifyAndEnableTwoFA)
			userTwoFARoute.POST("/disable", controller.DisableTwoFA)
		}
		apiRouter.POST("/user/2fa/verify", controller.VerifyLoginTwoFA)

		// ==================== 模型同步路由 ====================
		modelSyncRoute := apiRouter.Group("/admin/model-sync")
		modelSyncRoute.Use(middleware.AdminAuth())
		{
			modelSyncRoute.GET("/", controller.GetModelSyncStatus)
			modelSyncRoute.PUT("/", controller.SaveModelSyncSetting)
			modelSyncRoute.POST("/sync", controller.SyncModels)
			modelSyncRoute.GET("/aliases", controller.GetModelAliasList)
			modelSyncRoute.POST("/aliases", controller.UpsertModelAlias)
			modelSyncRoute.GET("/search", controller.SearchModels)
		}

		// ==================== 渠道上游更新检测路由 ====================
		upstreamRoute := apiRouter.Group("/admin/upstream")
		upstreamRoute.Use(middleware.AdminAuth())
		{
			upstreamRoute.GET("/", controller.GetUpstreamUpdateSetting)
			upstreamRoute.PUT("/", controller.SaveUpstreamUpdateSetting)
			upstreamRoute.POST("/check", controller.CheckUpstreamUpdates)
		}

		// ==================== 性能监控路由 ====================
		apiRouter.GET("/admin/performance", middleware.AdminAuth(), controller.GetPerformanceStats)
		apiRouter.GET("/metrics", controller.GetPrometheusMetrics)

		// ==================== 渠道亲和性路由 ====================
		channelAffinityRoute := apiRouter.Group("/admin/channel-affinity")
		channelAffinityRoute.Use(middleware.AdminAuth())
		{
			channelAffinityRoute.GET("/", controller.GetChannelAffinitySettings)
			channelAffinityRoute.PUT("/", controller.SaveChannelAffinitySettings)
			channelAffinityRoute.DELETE("/cache", controller.ClearChannelAffinityCache)
			channelAffinityRoute.GET("/cache/stats", controller.GetChannelAffinityCacheStatsHandler)
		}

		// ==================== 自定义 OAuth 提供商路由 ====================
		customOAuthRoute := apiRouter.Group("/admin/custom-oauth")
		customOAuthRoute.Use(middleware.AdminAuth())
		{
			customOAuthRoute.GET("/", controller.ListCustomOAuthProviders)
			customOAuthRoute.POST("/", controller.CreateCustomOAuthProvider)
			customOAuthRoute.PUT("/:id", controller.UpdateCustomOAuthProvider)
			customOAuthRoute.DELETE("/:id", controller.DeleteCustomOAuthProvider)
		}
		// 用户端 OAuth 路由（不需要认证）
		apiRouter.GET("/oauth/custom/:name", controller.CustomOAuthLogin)
		apiRouter.GET("/oauth/custom/:name/callback", controller.CustomOAuthCallback)
	}
}
