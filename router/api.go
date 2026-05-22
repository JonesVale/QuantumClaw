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
	apiRouter.Use(middleware.SecurityHeaders())
	{
		apiRouter.GET("/status", controller.GetStatus)
		apiRouter.GET("/setup/check", controller.CheckSetup)
		apiRouter.POST("/setup/complete", middleware.CriticalRateLimit(), controller.CompleteSetup)
		apiRouter.GET("/rss/articles", controller.GetRssArticles)
		apiRouter.GET("/models", middleware.UserAuth(), controller.DashboardListModels)
		apiRouter.GET("/models/rankings", middleware.UserAuth(), controller.ListModelRankings)
		apiRouter.POST("/fusion", middleware.UserAuth(), controller.HandleFusion)
		apiRouter.GET("/notice", controller.GetNotice)
		apiRouter.GET("/about", controller.GetAbout)
		apiRouter.GET("/home_page_content", controller.GetHomePageContent)

		// T_Languages API routes
		apiRouter.GET("/languages", controller.GetLanguages)
		apiRouter.GET("/translations", controller.GetTranslations)
		apiRouter.POST("/languages/seed", middleware.AdminAuth(), controller.SeedTranslations)
		apiRouter.POST("/languages/seed-public", controller.SeedTranslationsIfEmpty)
		controller.RegisterFreeChatRoutes(apiRouter)

		apiRouter.Any("/webhook/epay", middleware.WebhookIPWhitelist(), controller.EpayNotify)
		apiRouter.POST("/webhook/stripe", middleware.WebhookIPWhitelist(), controller.StripeWebhook)
		apiRouter.POST("/webhook/creem", middleware.WebhookIPWhitelist(), controller.CreemWebhook)
		apiRouter.POST("/webhook/waffo", middleware.WebhookIPWhitelist(), controller.WaffoWebhook)
		apiRouter.POST("/webhook/binance", middleware.WebhookIPWhitelist(), controller.BinanceWebhook)
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
				selfRoute.GET("/topup/list", controller.GetTopUpList)

				selfRoute.GET("/checkin", controller.GetCheckinStatus)
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
				selfRoute.GET("/balance", controller.GetSelfBalance)
				selfRoute.POST("/upgrade", controller.UpgradeToProvider)
				selfRoute.POST("/withdraw", controller.SubmitWithdrawal)
				selfRoute.GET("/withdraw/list", controller.GetMyWithdrawals)
				selfRoute.GET("/withdraw/available", controller.GetMyWithdrawable)
				selfRoute.GET("/withdraw/earnings", controller.GetMyEarningsByChannel)
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
			channelUserRoute.GET("/types", controller.GetChannelTypes)
			channelUserRoute.GET("/search", controller.SearchChannels)
			channelUserRoute.GET("/models", controller.ListAllModels)
			channelUserRoute.GET("/:id", controller.GetChannel)
			channelUserRoute.POST("/", controller.AddChannel)
			channelUserRoute.PUT("/", controller.UpdateChannel)
			channelUserRoute.DELETE("/:id", controller.DeleteChannel)
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
		/*
		  API routes registered at top of SetApiRouter:
		    apiRouter.GET("/languages", ...)
		    apiRouter.GET("/translations", ...)
		  Seed endpoint:
		    apiRouter.POST("/languages/seed", ...)
		*/

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
	}
}

