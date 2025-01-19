// routes/routes.go
package routes

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"smuggr.xyz/gatecam/api/v1/handlers"
	"smuggr.xyz/gatecam/common/config"

	"github.com/gin-gonic/gin"
)

var Config config.APIConfig

func logRequestDetails(c *gin.Context) {
	fmt.Println("========== Incoming Request ==========")
	fmt.Printf("Method: %s\n", c.Request.Method)
	fmt.Printf("URL: %s\n", c.Request.URL.String())
	fmt.Printf("Headers:\n")
	for key, values := range c.Request.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	if c.Request.Body != nil {
		bodyBytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Printf("Error reading body: %s\n", err)
		} else {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			fmt.Printf("Body: %s\n", string(bodyBytes))
		}
	}
	fmt.Println("======================================")
}

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
		externalCameraGroup.GET("/stream", handlers.HandleExternalCameraStream)
	}
}

func SetupDeviceRoutes(router *gin.Engine, externalRouter *gin.Engine, rootGroup *gin.RouterGroup, externalRootGroup *gin.RouterGroup) {
	devicesGroup := rootGroup.Group("/device")
	devicesGroup.Use(logRequestDetails)
	deviceGroup := devicesGroup.Group("/:id")
	{
		deviceGroup.Any("/*endpoint", handlers.HandleDeviceEndpoint)
	}

	externalDevicesGroup := externalRootGroup.Group("/device")
	externalDevicesGroup.Use(logRequestDetails)
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
	SetupDeviceRoutes(defaultRouter, externalRouter, rootGroup, externalRootGroup)

	handlers.Initialize()
}
