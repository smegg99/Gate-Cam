// api/v1/api.go
package v1

import (
	"fmt"
	"os"
	"strconv"

	"smuggr.xyz/gate-cam/api/v1/routes"
	"smuggr.xyz/gate-cam/common/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var DefaultRouter *gin.Engine
var ExternalRouter *gin.Engine
var Config *config.APIConfig

func Initialize() chan error {
	fmt.Println("initializing api/v1")

	Config = &config.Global.API
	gin.SetMode(os.Getenv("GIN_MODE"))

	// TODO: Add TLS support
	DefaultRouter = gin.Default()
	ExternalRouter = gin.Default()

	DefaultRouter.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:2137", "http://localhost:2138", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-Auth-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	ExternalRouter.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:2137", "http://localhost:2138", "http://gatecam.smuggr.xyz", "https://gatecam.smuggr.xyz"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-Auth-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	routes.Initialize(DefaultRouter, ExternalRouter)

	errCh := make(chan error)
	go func() {
		// err := DefaultRouter.RunTLS(":" + strconv.Itoa(int(Config.Port)), os.Getenv("TLS_CERT"), os.Getenv("TLS_KEY"))
		err := DefaultRouter.Run(":" + strconv.Itoa(int(Config.Port)))
		errCh <- err
	}()

	extErrCh := make(chan error)
	go func() {
		err := ExternalRouter.Run(":" + strconv.Itoa(int(Config.ExternalPort)))
		extErrCh <- err
	}()

	return errCh
}
