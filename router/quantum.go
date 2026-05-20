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
		quantumRouter.POST("/run", controller.QuantumRelay)
		quantumRouter.GET("/status/:task_id", controller.QuantumRelay)
		quantumRouter.POST("/cancel/:task_id", controller.QuantumRelay)
		quantumRouter.GET("/backends", controller.QuantumRelay)
	}
}
