package router

import (
	"github.com/quantumclaw/quantumclaw/controller"
	"github.com/quantumclaw/quantumclaw/middleware"

	"github.com/gin-gonic/gin"
)

func SetRelayRouter(router *gin.Engine) {
	router.Use(middleware.CORS())
	router.Use(middleware.GzipDecodeMiddleware())
	// https://platform.openai.com/docs/api-reference/introduction
	modelsRouter := router.Group("/v1/models")
	modelsRouter.Use(middleware.TokenAuth())
	{
		modelsRouter.GET("", controller.ListModels)
		modelsRouter.GET("/:model", controller.RetrieveModel)
	}
	relayV1Router := router.Group("/v1")
	relayV1Router.Use(middleware.RelayPanicRecover(), middleware.TokenAuth(), middleware.ModelRateLimit(), middleware.Distribute())
	{
		relayV1Router.Any("/oneapi/proxy/:channelid/*target", controller.Relay)
		relayV1Router.POST("/responses", controller.Relay)
		relayV1Router.POST("/completions", controller.Relay)
		relayV1Router.POST("/chat/completions", controller.Relay)
		relayV1Router.POST("/edits", controller.Relay)
		relayV1Router.POST("/images/generations", controller.Relay)
		relayV1Router.POST("/images/edits", controller.RelayNotImplemented)
		relayV1Router.POST("/images/variations", controller.RelayNotImplemented)
		relayV1Router.POST("/embeddings", controller.Relay)
		relayV1Router.POST("/engines/:model/embeddings", controller.Relay)
		relayV1Router.POST("/audio/transcriptions", controller.Relay)
		relayV1Router.POST("/audio/translations", controller.Relay)
		relayV1Router.POST("/audio/speech", controller.Relay)
		relayV1Router.GET("/files", controller.Relay)
		relayV1Router.POST("/files", controller.Relay)
		relayV1Router.DELETE("/files/:id", controller.Relay)
		relayV1Router.GET("/files/:id", controller.Relay)
		relayV1Router.GET("/files/:id/content", controller.Relay)
		relayV1Router.POST("/fine_tuning/jobs", controller.Relay)
		relayV1Router.GET("/fine_tuning/jobs", controller.Relay)
		relayV1Router.GET("/fine_tuning/jobs/:id", controller.Relay)
		relayV1Router.POST("/fine_tuning/jobs/:id/cancel", controller.Relay)
		relayV1Router.GET("/fine_tuning/jobs/:id/events", controller.Relay)
		relayV1Router.DELETE("/models/:model", controller.RelayNotImplemented)
		relayV1Router.POST("/moderations", controller.Relay)
		relayV1Router.POST("/assistants", controller.Relay)
		relayV1Router.GET("/assistants/:id", controller.Relay)
		relayV1Router.POST("/assistants/:id", controller.Relay)
		relayV1Router.DELETE("/assistants/:id", controller.Relay)
		relayV1Router.GET("/assistants", controller.Relay)
		relayV1Router.POST("/assistants/:id/files", controller.Relay)
		relayV1Router.GET("/assistants/:id/files/:fileId", controller.Relay)
		relayV1Router.DELETE("/assistants/:id/files/:fileId", controller.Relay)
		relayV1Router.GET("/assistants/:id/files", controller.Relay)
		relayV1Router.POST("/threads", controller.Relay)
		relayV1Router.GET("/threads/:id", controller.Relay)
		relayV1Router.POST("/threads/:id", controller.Relay)
		relayV1Router.DELETE("/threads/:id", controller.Relay)
		relayV1Router.POST("/threads/:id/messages", controller.Relay)
		relayV1Router.GET("/threads/:id/messages/:messageId", controller.Relay)
		relayV1Router.POST("/threads/:id/messages/:messageId", controller.Relay)
		relayV1Router.GET("/threads/:id/messages/:messageId/files/:filesId", controller.Relay)
		relayV1Router.GET("/threads/:id/messages/:messageId/files", controller.Relay)
		relayV1Router.POST("/threads/:id/runs", controller.Relay)
		relayV1Router.GET("/threads/:id/runs/:runsId", controller.Relay)
		relayV1Router.POST("/threads/:id/runs/:runsId", controller.Relay)
		relayV1Router.GET("/threads/:id/runs", controller.Relay)
		relayV1Router.POST("/threads/:id/runs/:runsId/submit_tool_outputs", controller.Relay)
		relayV1Router.POST("/threads/:id/runs/:runsId/cancel", controller.Relay)
		relayV1Router.GET("/threads/:id/runs/:runsId/steps/:stepId", controller.Relay)
		relayV1Router.GET("/threads/:id/runs/:runsId/steps", controller.Relay)

		// ==================== 异步任务 relay（Midjourney / 视频生成 / Suno）====================
		// 这些请求走异步任务系统：创建任务 → 返回 task_id → 后台处理 → 轮询状态
		relayV1Router.Any("/mj/*path", controller.Relay)
		relayV1Router.Any("/video/*path", controller.Relay)
		relayV1Router.Any("/suno/*path", controller.Relay)

		// ==================== Claude Messages API ====================
		relayV1Router.POST("/messages", controller.Relay)

		// ==================== OpenAI Batch API ====================
		relayV1Router.POST("/batches", controller.Relay)
		relayV1Router.GET("/batches/:id", controller.Relay)
		relayV1Router.POST("/batches/:id/cancel", controller.Relay)

		// ==================== OpenAI Vector Stores API ====================
		relayV1Router.POST("/vector_stores", controller.Relay)
		relayV1Router.GET("/vector_stores", controller.Relay)
		relayV1Router.GET("/vector_stores/:id", controller.Relay)
		relayV1Router.POST("/vector_stores/:id", controller.Relay)
		relayV1Router.DELETE("/vector_stores/:id", controller.Relay)
		relayV1Router.POST("/vector_stores/:id/file_batches", controller.Relay)
		relayV1Router.GET("/vector_stores/:id/file_batches/:batchId", controller.Relay)
		relayV1Router.POST("/vector_stores/:id/files", controller.Relay)
		relayV1Router.GET("/vector_stores/:id/files", controller.Relay)
		relayV1Router.DELETE("/vector_stores/:id/files/:fileId", controller.Relay)
		relayV1Router.GET("/vector_stores/:id/files/:fileId", controller.Relay)
	}
}
