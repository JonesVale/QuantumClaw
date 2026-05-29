package router



import (

	"net/http"



	"github.com/quantumclaw/quantumclaw/controller"

	"github.com/quantumclaw/quantumclaw/controller/auth"

	"github.com/quantumclaw/quantumclaw/docs"

	"github.com/quantumclaw/quantumclaw/middleware"



	"github.com/gin-contrib/gzip"

	"github.com/gin-gonic/gin"

)



const swaggerHTML = `<!DOCTYPE html>

<html lang="en">

<head>

  <meta charset="utf-8" />

  <meta name="viewport" content="width=device-width, initial-scale=1" />

  <title>QuantumClaw API Documentation</title>

  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />

</head>

<body>

  <div id="swagger-ui"></div>

  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>

  <script>

    SwaggerUIBundle({ url: "/api/swagger/doc.json", dom_id: "#swagger-ui" });

  </script>

</body>

</html>`



func SetApiRouter(router *gin.Engine) {

	apiRouter := router.Group("/api")

	apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))

	apiRouter.Use(middleware.GlobalAPIRateLimit())

	apiRouter.Use(middleware.SecurityHeaders())

	{

		apiRouter.GET("/status", controller.GetStatus)

		apiRouter.GET("/site-content", controller.GetSiteContent)

		apiRouter.GET("/site-stats", controller.GetSiteStats)

		apiRouter.GET("/site-features", controller.GetSiteFeatures)

		apiRouter.GET("/site-providers", controller.GetSiteProviders)

		apiRouter.GET("/setup/check", controller.CheckSetup)

		apiRouter.POST("/setup/complete", middleware.CriticalRateLimit(), controller.CompleteSetup)

		apiRouter.GET("/rss/articles", controller.GetRssArticles)

		apiRouter.GET("/models", controller.DashboardListModels)

		apiRouter.GET("/models/rankings", controller.ListModelRankings)

		apiRouter.GET("/model-catalog", controller.GetModelCatalog)

		apiRouter.GET("/brand-rankings", controller.GetBrandRankings)

		apiRouter.GET("/enterprise-clients", controller.ListEnterpriseClients)
		apiRouter.GET("/model-catalog/:model_name", controller.GetModelDetail)

		// 公开渠道类型查询（无需认证，前端 channels 页面需要）
		apiRouter.GET("/channel/types", controller.GetChannelTypes)

		apiRouter.POST("/models/sync", controller.SyncModelMetadata)

		apiRouter.GET("/models/seed-quantum", controller.SeedQuantumModels)

		apiRouter.POST("/fusion", middleware.UserAuth(), controller.HandleFusion)

		apiRouter.GET("/quantum/backends", middleware.UserAuth(), controller.GetQuantumBackends)

		apiRouter.GET("/quantum/providers", middleware.UserAuth(), controller.GetQuantumProviders)

		apiRouter.POST("/quantum/submit", middleware.UserAuth(), controller.SubmitQuantumTask)

		apiRouter.GET("/notice", controller.GetNotice)

		apiRouter.GET("/about", controller.GetAbout)

		apiRouter.GET("/home_page_content", controller.GetHomePageContent)



		controller.RegisterFreeChatRoutes(apiRouter)



		apiRouter.Any("/webhook/epay", middleware.WebhookIPWhitelist(), controller.EpayNotify)

		apiRouter.POST("/webhook/stripe", middleware.WebhookIPWhitelist(), controller.StripeWebhook)



	// Public FAQ, App Market routes

	apiRouter.GET("/faq", controller.GetFAQs)

	apiRouter.GET("/faq/categories", controller.GetFAQCategories)

	apiRouter.GET("/apps", controller.ListApps)

	apiRouter.GET("/apps/:id", controller.GetApp)



		apiRouter.POST("/webhook/creem", middleware.WebhookIPWhitelist(), controller.CreemWebhook)

		apiRouter.POST("/webhook/waffo", middleware.WebhookIPWhitelist(), controller.WaffoWebhook)

		apiRouter.POST("/webhook/binance", middleware.WebhookIPWhitelist(), controller.BinanceWebhook)

		apiRouter.Any("/webhook/alipay", middleware.WebhookIPWhitelist(), controller.AlipayNotify)

		apiRouter.Any("/webhook/worldfirst", middleware.WebhookIPWhitelist(), controller.WorldFirstWebhook)

		apiRouter.GET("/verification", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendEmailVerification)

		apiRouter.GET("/reset_password", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), controller.SendPasswordResetEmail)

		apiRouter.POST("/user/reset", middleware.CriticalRateLimit(), controller.ResetPassword)

		apiRouter.GET("/oauth/github", middleware.CriticalRateLimit(), auth.GitHubOAuth)

		apiRouter.GET("/oauth/oidc", middleware.CriticalRateLimit(), auth.OidcAuth)

		apiRouter.GET("/oauth/lark", middleware.CriticalRateLimit(), auth.LarkOAuth)

		apiRouter.GET("/oauth/state", middleware.CriticalRateLimit(), auth.GenerateOAuthCode)

		apiRouter.GET("/oauth/wechat", middleware.CriticalRateLimit(), auth.WeChatAuth)

		apiRouter.GET("/oauth/wechat/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), auth.WeChatBind)

		apiRouter.GET("/oauth/discord", middleware.CriticalRateLimit(), auth.DiscordOAuth)

		apiRouter.GET("/oauth/linuxdo", middleware.CriticalRateLimit(), auth.LinuxDoOAuth)

		apiRouter.GET("/oauth/discord/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), auth.DiscordBind)

		apiRouter.GET("/oauth/linuxdo/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), auth.LinuxDoBind)

		apiRouter.GET("/oauth/linuxdo/generate", middleware.CriticalRateLimit(), auth.GenerateLinuxDOAuthURL)

		apiRouter.GET("/oauth/telegram", middleware.CriticalRateLimit(), auth.TelegramOAuth)

		apiRouter.POST("/oauth/telegram", middleware.CriticalRateLimit(), auth.TelegramAuthHandler)

		apiRouter.POST("/oauth/telegram/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), auth.TelegramBindHandler)

		apiRouter.GET("/oauth/telegram/widget", middleware.CriticalRateLimit(), auth.GenerateTelegramWidgetOptions)

		apiRouter.GET("/oauth/email/bind", middleware.CriticalRateLimit(), middleware.UserAuth(), controller.EmailBind)



		apiRouter.POST("/webauthn/register/begin", middleware.CriticalRateLimit(), auth.WebAuthnBeginRegistration)

		apiRouter.POST("/webauthn/register/finish", middleware.CriticalRateLimit(), auth.WebAuthnFinishRegistration)

		apiRouter.POST("/webauthn/login/begin", middleware.CriticalRateLimit(), auth.WebAuthnBeginAuthentication)

		apiRouter.POST("/webauthn/login/finish", middleware.CriticalRateLimit(), auth.WebAuthnFinishAuthentication)

		apiRouter.POST("/topup", middleware.AdminAuth(), middleware.RequirePaymentAuth(), controller.AdminTopUp)



		userRoute := apiRouter.Group("/user")

		{

			userRoute.POST("/register", middleware.CriticalRateLimit(), middleware.LoginRateLimit(), middleware.TurnstileCheck(), controller.Register)

			userRoute.POST("/login", middleware.CriticalRateLimit(), middleware.LoginRateLimit(), controller.Login)

			userRoute.GET("/logout", controller.Logout)



			// Password management routes

			userRoute.POST("/password/change", middleware.UserAuth(), controller.ChangePassword)

			userRoute.POST("/password/admin_reset", middleware.AdminAuth(), controller.AdminResetUserPassword)

			// Admin/info endpoint

			userRoute.GET("/info", middleware.UserAuth(), controller.GetUserInfo)



			selfRoute := userRoute.Group("/self")

			selfRoute.Use(middleware.UserAuth())

			{

				selfRoute.GET("/dashboard", controller.GetUserDashboard)

				selfRoute.GET("/", controller.GetSelf)

				selfRoute.PUT("/", controller.UpdateSelf)

				selfRoute.DELETE("/", controller.DeleteSelf)

				selfRoute.GET("/token", controller.GenerateAccessToken)

				selfRoute.GET("/aff", controller.GetAffCode)

				selfRoute.POST("/topup", controller.TopUp)

				selfRoute.GET("/available_models", controller.GetUserAvailableModels)



				selfRoute.GET("/topup/info", controller.GetTopUpInfo)

				selfRoute.POST("/topup/epay", controller.RequestEpayTopUp)

				selfRoute.POST("/topup/stripe", controller.RequestStripeTopUp)

				selfRoute.POST("/topup/creem", controller.RequestCreemTopUp)

				selfRoute.POST("/topup/waffo", controller.RequestWaffoTopUp)

				selfRoute.POST("/topup/binance", controller.RequestBinanceTopUp)

				selfRoute.POST("/topup/alipay", controller.RequestAlipayTopUp)

				selfRoute.POST("/topup/worldfirst", controller.RequestWorldFirstTopUp)

				selfRoute.GET("/topup/list", controller.GetTopUpList)



				selfRoute.GET("/checkin", controller.GetCheckinStatus)

				selfRoute.GET("/checkin/history", controller.GetCheckinHistory)

				selfRoute.POST("/checkin", controller.DoCheckin)



				selfRoute.GET("/subscription/plans", controller.GetSubscriptionPlans)

				selfRoute.GET("/subscription/self", controller.GetSubscriptionSelf)



				selfRoute.GET("/webauthn/credentials", auth.WebAuthnGetCredentials)

				selfRoute.DELETE("/webauthn/credentials/:id", auth.WebAuthnDeleteCredential)



				selfRoute.GET("/transaction_logs", controller.GetTransactionLogs)

				selfRoute.GET("/security/activity", controller.GetSecurityActivity)



				selfRoute.GET("/notifications", controller.GetNotifications)

				selfRoute.GET("/notifications/unread_count", controller.GetUnreadNotificationCount)

				selfRoute.PUT("/notifications/:id/read", controller.MarkNotificationRead)

				selfRoute.PUT("/notifications/read_all", controller.MarkAllNotificationsRead)

				selfRoute.GET("/billing/stats", controller.GetBillingStats)

				selfRoute.GET("/billing/records", controller.GetBillingRecords)

				selfRoute.GET("/balance", controller.GetSelfBalance)

				selfRoute.GET("/team", controller.GetMyTeam)

	selfRoute.POST("/upgrade", controller.UpgradeToProvider)

				selfRoute.POST("/withdraw", controller.SubmitWithdrawal)

				selfRoute.GET("/withdraw/list", controller.GetMyWithdrawals)

				selfRoute.GET("/withdraw/available", controller.GetMyWithdrawable)

				selfRoute.GET("/withdraw/earnings", controller.GetMyEarningsByChannel)



			// User feedback, inference-nodes, apps

			selfRoute.GET("/feedback", controller.GetMyFeedback)

			selfRoute.POST("/feedback", middleware.CriticalRateLimit(), controller.SubmitFeedback)



			selfRoute.POST("/inference-nodes", middleware.CriticalRateLimit(), controller.CreateInferenceNode)

			selfRoute.GET("/inference-nodes", controller.ListInferenceNodes)

			selfRoute.PUT("/inference-nodes/:id", controller.UpdateInferenceNode)

			selfRoute.DELETE("/inference-nodes/:id", controller.DeleteInferenceNode)

			selfRoute.POST("/inference-nodes/:id/test", middleware.CriticalRateLimit(), controller.TestInferenceNode)



			selfRoute.POST("/apps", middleware.CriticalRateLimit(), controller.SubmitApp)

			selfRoute.GET("/apps", controller.GetMyApps)



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

				adminRoute.POST("/add_balance", controller.AdminAddBalance)

				adminRoute.GET("/balance/:id", controller.GetUserBalanceByAdmin)

				adminRoute.GET("/withdrawals", controller.AdminGetWithdrawals)

				adminRoute.POST("/withdrawals/:id/approve", controller.AdminApproveWithdrawal)

				adminRoute.POST("/withdrawals/:id/reject", controller.AdminRejectWithdrawal)

				adminRoute.POST("/withdrawals/:id/complete", controller.AdminCompleteWithdrawal)

			}

		}

		optionRoute := apiRouter.Group("/option")

		optionRoute.Use(middleware.RootAuth())

		{

			optionRoute.GET("/", controller.GetOptions)

			optionRoute.PUT("/", controller.UpdateOption)

		}

		// 供应商渠道管理（所有登录用户可管理自己的渠道）

		channelUserRoute := apiRouter.Group("/channel")

		channelUserRoute.Use(middleware.UserAuth())

		{

			channelUserRoute.GET("/", controller.GetAllChannels)

			channelUserRoute.GET("/search", controller.SearchChannels)

			channelUserRoute.GET("/models", controller.ListAllModels)

			channelUserRoute.GET("/:id", controller.GetChannel)

			channelUserRoute.POST("/", controller.AddChannel)

			channelUserRoute.PUT("/", controller.UpdateChannel)

			channelUserRoute.DELETE("/:id", controller.DeleteChannel)

		}

		// Menu permission routes (public - filtered by user role, guest-friendly)

		apiRouter.GET("/menus", controller.GetMenus)



		// Promo ads (public)

		apiRouter.GET("/promo-ads", controller.GetPromoAds)



		// Promo ads (admin only)

		adminPromoRoute := apiRouter.Group("/admin/promo-ads")

		adminPromoRoute.Use(middleware.AdminAuth())

		{

			adminPromoRoute.GET("/", controller.AdminGetAllPromoAds)

			adminPromoRoute.POST("/", controller.AdminCreatePromoAd)

			adminPromoRoute.PUT("/", controller.AdminUpdatePromoAd)

			adminPromoRoute.DELETE("/:id", controller.AdminDeletePromoAd)

		}



		// Menu permission routes (admin only)

		adminMenuRoute := apiRouter.Group("/admin/menus")

		adminMenuRoute.Use(middleware.AdminAuth())

		{

			adminMenuRoute.GET("/", controller.AdminGetAllMenus)

			adminMenuRoute.POST("/", controller.AdminCreateOrUpdateMenu)

			adminMenuRoute.DELETE("/:id", controller.AdminDeleteMenu)

		}



		// 渠道管理维护（管理员）

		channelAdminRoute := apiRouter.Group("/channel")

		channelAdminRoute.Use(middleware.AdminAuth())

		{

			channelAdminRoute.GET("/test", controller.TestChannels)

			channelAdminRoute.GET("/test/:id", controller.TestChannel)

			channelAdminRoute.GET("/update_balance", controller.UpdateAllChannelsBalance)

			channelAdminRoute.GET("/update_balance/:id", controller.UpdateChannelBalance)

			channelAdminRoute.GET("/profit", controller.GetChannelProfit)

			channelAdminRoute.POST("/pricing", controller.SetChannelPricing)

			channelAdminRoute.POST("/category", controller.SetChannelCategory)

			channelAdminRoute.DELETE("/disabled", controller.DeleteDisabledChannel)

		}

		tokenRoute := apiRouter.Group("/token")

		{

			tokenRoute.GET("/query", controller.QueryTokenByKey)

		}

		tokenRoute.Use(middleware.UserAuth())

		{

			tokenRoute.GET("/", controller.GetAllTokens)

			tokenRoute.GET("/search", controller.SearchTokens)

			tokenRoute.GET("/:id", controller.GetToken)

			tokenRoute.POST("/", controller.AddToken)

			tokenRoute.PUT("/", controller.UpdateToken)

			tokenRoute.DELETE("/:id", controller.DeleteToken)



		// Commission / Affiliate routes

		commissionRoute := apiRouter.Group("/commission")

		commissionRoute.Use(middleware.AdminAuth())

		{

			commissionRoute.GET("/setting", controller.GetCommissionSetting)

			commissionRoute.PUT("/setting", controller.SaveCommissionSetting)

			commissionRoute.GET("/withdrawals", controller.AdminGetWithdrawals)

			commissionRoute.PUT("/withdrawals/:id/process", controller.AdminProcessWithdrawal)

		}

		selfCommRoute := apiRouter.Group("/commission")

		selfCommRoute.Use(middleware.UserAuth())

		{

			selfCommRoute.GET("/self/records", controller.GetMyCommissionRecords)

			selfCommRoute.GET("/self/withdrawals", controller.GetMyWithdrawals)

			selfCommRoute.POST("/self/withdraw", controller.RequestWithdrawal)

		}

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



		subscriptionRoute := apiRouter.Group("/subscription")

		{

			subscriptionRoute.Use(middleware.UserAuth())

			subscriptionRoute.GET("/plans", controller.GetSubscriptionPlans)

			subscriptionRoute.GET("/self", controller.GetSubscriptionSelf)

		}



		// ── Sub2API 订阅转 API ──

		sub2Route := apiRouter.Group("/sub2api")

		sub2Route.Use(middleware.UserAuth())

		{

			sub2Route.GET("/providers", controller.GetSub2Providers)

			sub2Route.GET("/credentials", controller.ListSub2Credentials)

			sub2Route.POST("/credentials", controller.AddSub2Credential)

			sub2Route.PUT("/credentials/:id", controller.UpdateSub2Credential)

			sub2Route.DELETE("/credentials/:id", controller.DeleteSub2Credential)

			sub2Route.GET("/credentials/:id/test", controller.TestSub2Credential)

		}



		// Admin Sub2API

		adminSub2Route := apiRouter.Group("/admin/sub2api")

		adminSub2Route.Use(middleware.AdminAuth())

		{

			adminSub2Route.GET("/credentials", controller.AdminListAllSub2Credentials)

			adminSub2Route.DELETE("/credentials/:id", controller.AdminDeleteSub2Credential)



			// ── Schema 管理 ──

			adminSub2Route.GET("/schemas", controller.ListSub2APISchemas)

			adminSub2Route.GET("/schemas/:id", controller.GetSub2APISchema)

			adminSub2Route.POST("/schemas", controller.CreateSub2APISchema)

			adminSub2Route.PUT("/schemas/:id", controller.UpdateSub2APISchema)

			adminSub2Route.DELETE("/schemas/:id", controller.DeleteSub2APISchema)

			adminSub2Route.GET("/schemas/:id/test", controller.TestSub2APISchema)



			// ── 健康监控 ──

			adminSub2Route.GET("/health", controller.GetSub2APISchemaHealth)

			adminSub2Route.POST("/health/check", controller.TriggerSub2APIHealthCheck)

		}



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



		apiRouter.POST("/webhook/waffo_pancake", controller.HandleWaffoPancakeWebhook)



		userTwoFARoute := apiRouter.Group("/user/2fa")

		userTwoFARoute.Use(middleware.UserAuth())

		{

			userTwoFARoute.GET("/", controller.GetTwoFAStatus)

			userTwoFARoute.POST("/init", controller.InitTwoFA)

			userTwoFARoute.POST("/enable", controller.VerifyAndEnableTwoFA)

			userTwoFARoute.POST("/disable", controller.DisableTwoFA)

		}

		apiRouter.POST("/user/2fa/verify", controller.VerifyLoginTwoFA)



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



		upstreamRoute := apiRouter.Group("/admin/upstream")

		upstreamRoute.Use(middleware.AdminAuth())

		{

			upstreamRoute.GET("/", controller.GetUpstreamUpdateSetting)

			upstreamRoute.PUT("/", controller.SaveUpstreamUpdateSetting)

			upstreamRoute.POST("/check", controller.CheckUpstreamUpdates)

		}



		apiRouter.GET("/admin/performance", middleware.AdminAuth(), controller.GetPerformanceStats)

		apiRouter.GET("/admin/monitor", middleware.AdminAuth(), controller.GetAdminMonitor)

		apiRouter.GET("/metrics", controller.GetPrometheusMetrics)



		channelAffinityRoute := apiRouter.Group("/admin/channel-affinity")

		channelAffinityRoute.Use(middleware.AdminAuth())

		{

			channelAffinityRoute.GET("/", controller.GetChannelAffinitySettings)

			channelAffinityRoute.PUT("/", controller.SaveChannelAffinitySettings)

			channelAffinityRoute.DELETE("/cache", controller.ClearChannelAffinityCache)

			channelAffinityRoute.GET("/cache/stats", controller.GetChannelAffinityCacheStatsHandler)

		}



		customOAuthRoute := apiRouter.Group("/admin/custom-oauth")

		customOAuthRoute.Use(middleware.AdminAuth())

		{

			customOAuthRoute.GET("/", controller.ListCustomOAuthProviders)

			customOAuthRoute.POST("/", controller.CreateCustomOAuthProvider)

			customOAuthRoute.PUT("/:id", controller.UpdateCustomOAuthProvider)

			customOAuthRoute.DELETE("/:id", controller.DeleteCustomOAuthProvider)

		}

		apiRouter.GET("/oauth/custom/:name", controller.CustomOAuthLogin)

		apiRouter.GET("/oauth/custom/:name/callback", controller.CustomOAuthCallback)



		taskRoute := apiRouter.Group("/task")

		taskRoute.Use(middleware.UserAuth())

		{

			taskRoute.POST("/midjourney", controller.CreateMidjourneyTask)

			taskRoute.GET("/midjourney/:task_id", controller.GetMidjourneyTask)

			taskRoute.POST("/video", controller.CreateVideoTask)

			taskRoute.POST("/suno", controller.CreateSunoTask)

			taskRoute.GET("/:task_id", controller.GetTaskStatus)

			taskRoute.GET("/", controller.ListUserTasks)

			taskRoute.POST("/:task_id/cancel", controller.CancelTask)

			taskRoute.DELETE("/:task_id", controller.DeleteTask)

		}



		adminTaskRoute := apiRouter.Group("/admin/task")

		adminTaskRoute.Use(middleware.AdminAuth())

		{

			adminTaskRoute.GET("/", controller.AdminGetAllTasks)

			adminTaskRoute.POST("/poll", controller.AdminPollTasks)

		}



		// Password management routes

		passwordRoute := apiRouter.Group("/password")

		passwordRoute.Use(middleware.UserAuth())

		{

			passwordRoute.POST("/change", controller.ChangePassword)

		}

		adminPasswordRoute := apiRouter.Group("/password")

		adminPasswordRoute.Use(middleware.AdminAuth())

		{

			adminPasswordRoute.POST("/reset-user", controller.AdminResetUserPassword)

		}

		// Emergency password reset (requires env token, no session needed)

		apiRouter.POST("/password/emergency-reset", middleware.CriticalRateLimit(), controller.EmergencyPasswordReset)

		// Distributor routes

		distributorRoute := apiRouter.Group("/distributor")

		distributorRoute.Use(middleware.AdminAuth())

		{

			distributorRoute.GET("/", controller.GetDistributors)

			distributorRoute.POST("/", controller.CreateDistributor)

			distributorRoute.PUT("/:id", controller.UpdateDistributor)

			distributorRoute.GET("/:id/pricing", controller.GetDistributorPricing)

			distributorRoute.PUT("/:id/pricing", controller.SetDistributorPricing)

			distributorRoute.GET("/:id/revenue", controller.GetDistributorRevenue)

		}

		// Distributor self-service routes

		apiRouter.GET("/distributor/self", middleware.UserAuth(), controller.GetMyDistributor)



		// Settlement config routes

		settlementRoute := apiRouter.Group("/settlement")

		settlementRoute.Use(middleware.AdminAuth())

		{

			settlementRoute.GET("/config", controller.GetSettlementConfigs)

			settlementRoute.POST("/config", controller.CreateSettlementConfig)

			settlementRoute.PUT("/config/:id", controller.UpdateSettlementConfig)

			settlementRoute.DELETE("/config/:id", controller.DeleteSettlementConfig)

		}



		// Hourly settlement report routes
		apiRouter.GET("/settlement/hourly", middleware.AdminAuth(), controller.GetHourlySettlements)

		// Transaction routes

		apiRouter.GET("/transactions", middleware.UserAuth(), controller.GetTransactions)



		// Platform config routes

		apiRouter.GET("/platform/config", middleware.AdminAuth(), controller.GetPlatformConfigs)

		apiRouter.PUT("/platform/config", middleware.AdminAuth(), controller.UpdatePlatformConfig)



		// Reseller self-service routes

		apiRouter.GET("/reseller/balance", middleware.UserAuth(), controller.GetResellerBalance)

		apiRouter.POST("/reseller/withdraw", middleware.UserAuth(), controller.SubmitResellerWithdrawal)

		apiRouter.GET("/reseller/stats", middleware.UserAuth(), controller.GetResellerStats)



		// Admin reseller management

		apiRouter.GET("/admin/resellers", middleware.AdminAuth(), controller.ListResellers)

		apiRouter.GET("/admin/withdrawals", middleware.AdminAuth(), controller.ListWithdrawals)

		apiRouter.POST("/admin/withdrawals/:id/approve", middleware.AdminAuth(), controller.ApproveWithdrawal)



		// ── 定价审核与预览（管理员）──

		apiRouter.GET("/admin/billing/audit", middleware.AdminAuth(), controller.GetBillingAudit)

		apiRouter.POST("/admin/billing/preview", middleware.AdminAuth(), controller.PreviewBilling)



		// ── 订阅阶梯定价（管理员）──

		apiRouter.GET("/admin/subscription/:id/tiers", middleware.AdminAuth(), controller.GetSubscriptionTierInfo)



		// ── 联网搜索配置（管理员）──

		apiRouter.GET("/search/config", middleware.AdminAuth(), controller.GetSearchConfig)

		apiRouter.PUT("/search/config", middleware.AdminAuth(), controller.UpdateSearchConfig)

		apiRouter.POST("/search/test", middleware.AdminAuth(), controller.TestSearch)

		apiRouter.GET("/search/providers", middleware.AdminAuth(), controller.GetSearchProviders)



		// ── Geo 地理服务配置（管理员）──

		apiRouter.GET("/geo/config", middleware.AdminAuth(), controller.GetGeoConfig)

		apiRouter.PUT("/geo/config", middleware.AdminAuth(), controller.UpdateGeoConfig)

		apiRouter.POST("/geo/test", middleware.AdminAuth(), controller.TestGeo)

		apiRouter.GET("/geo/providers", middleware.AdminAuth(), controller.GetGeoProviders)



		// ── Geo 地理服务（用户端）──

		apiRouter.GET("/geo/query", controller.GeoRedirectHandler)



		// ── 智能路由配置（管理员）──

		apiRouter.GET("/router/config", middleware.AdminAuth(), controller.GetRouterConfig)

		apiRouter.PUT("/router/config", middleware.AdminAuth(), controller.UpdateRouterConfig)

		apiRouter.GET("/router/performance", middleware.AdminAuth(), controller.GetRouterPerformance)

		apiRouter.POST("/router/performance/reset", middleware.AdminAuth(), controller.ResetRouterPerformance)



		// ── API 文档（Swagger）──

		apiRouter.GET("/swagger/doc.json", func(c *gin.Context) {

			c.Data(http.StatusOK, "application/json", docs.SwaggerJSON)

		})

		apiRouter.GET("/swagger", func(c *gin.Context) {

			c.Redirect(http.StatusMovedPermanently, "/api/swagger/index.html")

		})

		apiRouter.GET("/swagger/index.html", func(c *gin.Context) {

			c.Header("Content-Type", "text/html; charset=utf-8")

			c.String(http.StatusOK, swaggerHTML)

		})



		// ── 级联架构（子节点注册 + 管理）──

		apiRouter.POST("/cascade/node/register", middleware.CriticalRateLimit(), controller.CascadeRegister)



	// ==================== Model Brands (Admin) ====================

	apiRouter.GET("/admin/model-brands", middleware.AdminAuth(), controller.GetModelBrands)

	apiRouter.POST("/admin/model-brands/configure", middleware.CriticalRateLimit(), middleware.AdminAuth(), controller.ConfigureModelBrand)
apiRouter.POST("/admin/model-brands/sync-all", middleware.CriticalRateLimit(), middleware.AdminAuth(), controller.SyncAllBrandModels)
apiRouter.POST("/admin/model-brands/configure-all", middleware.CriticalRateLimit(), middleware.AdminAuth(), controller.ConfigureAllBrands)

		apiRouter.POST("/cascade/node/heartbeat", middleware.CascadeAuth(), controller.CascadeHeartbeat)

		apiRouter.GET("/cascade/tokens/sync", middleware.CascadeAuth(), controller.CascadeTokenSync)

		apiRouter.POST("/cascade/users/batch", middleware.CascadeAuth(), controller.CascadeUserBatch)

		apiRouter.POST("/cascade/billing/batch", middleware.CascadeAuth(), controller.CascadeBillingBatch)

		apiRouter.GET("/cascade/config", middleware.CascadeAuth(), controller.CascadeConfigSync)





	// Admin: Feedback, FAQ, Apps, Inference Nodes

	apiRouter.GET("/admin/feedback", middleware.AdminAuth(), controller.ListFeedback)

	apiRouter.POST("/admin/feedback/:id/respond", middleware.AdminAuth(), controller.RespondFeedback)

	apiRouter.POST("/admin/feedback/:id/status", middleware.AdminAuth(), controller.UpdateFeedbackStatus)

	apiRouter.POST("/admin/faq", middleware.CriticalRateLimit(), middleware.AdminAuth(), controller.CreateFAQ)

	apiRouter.PUT("/admin/faq/:id", middleware.AdminAuth(), controller.UpdateFAQ)

	apiRouter.DELETE("/admin/faq/:id", middleware.AdminAuth(), controller.DeleteFAQ)

	apiRouter.GET("/admin/apps", middleware.AdminAuth(), controller.AdminListApps)

	apiRouter.POST("/admin/apps/:id/status", middleware.AdminAuth(), controller.AdminUpdateAppStatus)
	apiRouter.POST("/admin/apps/sync", middleware.AdminAuth(), controller.AdminSyncPopularApps)

	apiRouter.GET("/admin/inference-nodes", middleware.AdminAuth(), controller.AdminListInferenceNodes)

		apiRouter.GET("/cascade/nodes", middleware.AdminAuth(), controller.CascadeListNodes)

		apiRouter.DELETE("/cascade/nodes/:id", middleware.AdminAuth(), controller.CascadeDeleteNode)

	}

}



