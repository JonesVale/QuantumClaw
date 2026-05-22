package router

import (
	"github.com/quantumclaw/quantumclaw/controller"
	"github.com/quantumclaw/quantumclaw/middleware"

	"github.com/gin-gonic/gin"
)

// SetQuantumRouter 注册量子算力路由
func SetQuantumRouter(router *gin.Engine) {
	quantumRouter := router.Group("/v1/quantum")
	quantumRouter.Use(
		middleware.RelayPanicRecover(),
		middleware.TokenAuth(),
		middleware.Distribute(),
	)
	{
		quantumRouter.POST("/tasks", controller.SubmitQuantumTask)
		quantumRouter.GET("/backends", controller.GetQuantumBackends)
		quantumRouter.GET("/providers", controller.GetQuantumProviders)
	}
}
