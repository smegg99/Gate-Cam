// routes/routes.go
package routes

import (
	"smuggr.xyz/gatecam/api/v1/handlers"
	"smuggr.xyz/gatecam/common/config"

	"github.com/gin-gonic/gin"
)

var Config config.APIConfig

func SetupCameraRoutes(router *gin.Engine, externalRouter *gin.Engine, rootGroup *gin.RouterGroup, externalRootGroup *gin.RouterGroup) {
	camerasGroup := rootGroup.Group("/camera")
	cameraGroup := camerasGroup.Group("/:id")
	{
		cameraGroup.GET("/stream", handlers.HandleCameraStream)
		cameraGroup.GET("/raw_grayscale_frame", handlers.HandleCameraGrayscaleFrame)
		cameraGroup.GET("/raw_color_frame", handlers.HandleCameraColorFrame)
	}

	externalCamerasGroup := externalRootGroup.Group("/camera")
	externalCameraGroup := externalCamerasGroup.Group("/:id")
	{
		externalCameraGroup.GET("/stream", handlers.HandleCameraStream)
	}

	devicesGroup := rootGroup.Group("/device")
	deviceGroup := devicesGroup.Group("/:id")
	{
		deviceGroup.Any("/*endpoint", handlers.HandleDeviceEndpoint)
	}

	externalDevicesGroup := externalRootGroup.Group("/device")
	externalDeviceGroup := externalDevicesGroup.Group("/:id")
	{
		externalDeviceGroup.Any("/*endpoint", handlers.HandleExternalDeviceEndpoint)
	}
}

func Initialize(defaultRouter *gin.Engine, externalRouter *gin.Engine) {
	Config = config.Global.API

	rootGroup := defaultRouter.Group("/api/v1")
	externalRootGroup := externalRouter.Group("/api/v1")

	SetupCameraRoutes(defaultRouter, externalRouter, rootGroup, externalRootGroup)

	handlers.Initialize()
}
