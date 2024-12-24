// handlers/handlers.go
package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"smuggr.xyz/gate-cam/common/config"
	"smuggr.xyz/gate-cam/core/cameras"

	"github.com/gin-gonic/gin"
)

var Config *config.APIConfig

func Respond(c *gin.Context, code int, data interface{}) {
	accept := c.GetHeader("Accept")
	switch {
	case strings.Contains(accept, "application/json"):
		fallthrough
	default:
		c.JSON(code, data)
	}
}

func HandleCameraStream(c *gin.Context) {
	camID := c.Param("id")

	cam, ok := cameras.Server.GetCamera(camID)
	if !ok {
		Respond(c, http.StatusNotFound, gin.H{"error": fmt.Sprintf("camera not found: %s", camID)})
		return
	}

	c.Writer.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")

	for {
		frame, err := cam.ReadFrame()
		if err != nil {
			fmt.Printf("error reading frame from camera %s: %v\n", camID, err)
			break
		}

		fmt.Fprintf(c.Writer, "--frame\r\n")
		fmt.Fprintf(c.Writer, "Content-Type: image/jpeg\r\n")
		fmt.Fprintf(c.Writer, "Content-Length: %d\r\n\r\n", len(frame))
		_, err = c.Writer.Write(frame)
		if err != nil {
			fmt.Printf("client disconnected from camera %s: %v\n", camID, err)
			break
		}

		if flusher, ok := c.Writer.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}

func Initialize() {
	fmt.Println("initializing handlers")
	Config = &config.Global.API
}
