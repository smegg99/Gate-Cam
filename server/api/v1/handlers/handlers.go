// handlers/handlers.go
package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"smuggr.xyz/gatecam/common/config"
	"smuggr.xyz/gatecam/core/cameras"
	"smuggr.xyz/gatecam/core/devices"

	"github.com/gin-gonic/gin"
)

var Config *config.APIConfig

func handleDeviceEndpoint(c *gin.Context, device *devices.Device) {
    endpoint := c.Param("endpoint")
    targetURL := fmt.Sprintf("http://%s:%d%s", device.GetIP(), device.GetPort(), endpoint)

    bodyBytes, err := io.ReadAll(c.Request.Body)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
        return
    }

    c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

    req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(bodyBytes))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
        return
    }

    for key, values := range c.Request.Header {
        for _, value := range values {
            req.Header.Add(key, value)
        }
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to forward request", "details": err.Error()})
        return
    }
    defer resp.Body.Close()

    fmt.Printf("Forwarded request to %s, response status: %d\n", targetURL, resp.StatusCode)

    for key, values := range resp.Header {
        for _, value := range values {
            c.Writer.Header().Add(key, value)
        }
    }
    c.Writer.WriteHeader(resp.StatusCode)
    io.Copy(c.Writer, resp.Body)
}

func handleCameraStream(c *gin.Context, cam *cameras.Camera) {
    c.Writer.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")

    for {
        frame, err := cam.ReadFrame(config.ModeJPEGStream)
        if err != nil {
            fmt.Printf("error reading frame from camera %s: %v\n", cam.Name, err)
            break
        }

        fmt.Fprintf(c.Writer, "--frame\r\n")
        fmt.Fprintf(c.Writer, "Content-Type: image/jpeg\r\n")
        fmt.Fprintf(c.Writer, "Content-Length: %d\r\n\r\n", len(frame))
        _, err = c.Writer.Write(frame)
        if err != nil {
            fmt.Printf("client disconnected from camera %s: %v\n", cam.Name, err)
            break
        }

        if flusher, ok := c.Writer.(http.Flusher); ok {
            flusher.Flush()
        }
    }
}

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

	handleCameraStream(c, cam)
}

func HandleExternalCameraStream(c *gin.Context) {
	camID := c.Param("id")

	cam, ok := cameras.Server.GetCamera(camID)
	if !ok {
		Respond(c, http.StatusNotFound, gin.H{"error": fmt.Sprintf("camera not found: %s", camID)})
		return
	}

	user, pass, ok := c.Request.BasicAuth()
	if !ok || user != cam.Name || pass != cam.GetAccessKey() {
		c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	handleCameraStream(c, cam)
}

func HandleCameraGrayscaleFrame(c *gin.Context) {
	camID := c.Param("id")

	cam, ok := cameras.Server.GetCamera(camID)
	if !ok {
		Respond(c, http.StatusNotFound, gin.H{"error": fmt.Sprintf("camera not found: %s", camID)})
		return
	}

	frame, err := cam.ReadFrame(config.ModeGrayscaleFrame)
	if err != nil {
		fmt.Printf("error capturing grayscale frame from camera %s: %v\n", camID, err)
		Respond(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(frame)))
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Cache-Control", "cache")
	c.Writer.Header().Set("Pragma", "cache")

	_, err = c.Writer.Write(frame)
	if err != nil {
		fmt.Printf("error sending frame to client for camera %s: %v\n", camID, err)
		return
	}

	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

func HandleCameraColorFrame(c *gin.Context) {
	camID := c.Param("id")

	cam, ok := cameras.Server.GetCamera(camID)
	if !ok {
		Respond(c, http.StatusNotFound, gin.H{"error": fmt.Sprintf("camera not found: %s", camID)})
		return
	}

	frame, err := cam.ReadFrame(config.ModeColorFrame)
	if err != nil {
		fmt.Printf("error capturing color frame from camera %s: %v\n", camID, err)
		Respond(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(frame)))
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Cache-Control", "cache")
	c.Writer.Header().Set("Pragma", "cache")

	_, err = c.Writer.Write(frame)
	if err != nil {
		fmt.Printf("error sending frame to client for camera %s: %v\n", camID, err)
		return
	}

	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

func HandleExternalDeviceEndpoint(c *gin.Context) {
    devID := c.Param("id")
	device, ok := devices.Server.GetDevice(devID)
	if !ok {
		Respond(c, http.StatusNotFound, gin.H{"error": fmt.Sprintf("device not found: %s", devID)})
		return
	}

    user, pass, ok := c.Request.BasicAuth()
	if !ok || user != device.Name || pass != device.GetAccessKey() {
		c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

    handleDeviceEndpoint(c, device)
}

func HandleDeviceEndpoint(c *gin.Context) {
    devID := c.Param("id")
	device, ok := devices.Server.GetDevice(devID)
	if !ok {
		Respond(c, http.StatusNotFound, gin.H{"error": fmt.Sprintf("device not found: %s", devID)})
		return
	}

    handleDeviceEndpoint(c, device)
}

func Initialize() {
	fmt.Println("initializing handlers")
	Config = &config.Global.API
}
