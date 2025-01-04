// routes/routes.go
package routes

import (
	"smuggr.xyz/gate-cam/api/v1/handlers"
	"smuggr.xyz/gate-cam/common/config"

	"github.com/gin-gonic/gin"
)

var Config config.APIConfig

func SetupCameraRoutes(router *gin.Engine, rootGroup *gin.RouterGroup) {
	camerasGroup := rootGroup.Group("/cameras")
	{
		// camerasGroup.GET("", handlers.GetCameras)
	}

	cameraGroup := camerasGroup.Group("/:id")
	{
		// cameraGroup.GET("", handlers.GetCamera)
		cameraGroup.GET("/stream", handlers.HandleCameraStream)
		cameraGroup.GET("/raw_grayscale_frame", handlers.HandleCameraGrayscaleFrame)
		cameraGroup.GET("/raw_color_frame", handlers.HandleCameraColorFrame)
	}
}

func Initialize(defaultRouter *gin.Engine) {
	Config = config.Global.API

	rootGroup := defaultRouter.Group("/api/v1")

	SetupCameraRoutes(defaultRouter, rootGroup)

	handlers.Initialize()
}
