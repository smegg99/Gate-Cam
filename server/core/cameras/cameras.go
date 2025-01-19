// core/cameras/cameras.go
package cameras

import (
	"fmt"

	"smuggr.xyz/gatecam/common/config"
)

var Config config.GlobalConfig
var Server *MultiCamServer

func loadCameras() {
	for _, camConfig := range Config.Cameras {
		cam, err := NewCamera(camConfig)
		if err != nil {
			fmt.Printf("error creating camera %s: %v\n", camConfig.Name, err)
			continue
		}
		Server.AddCamera(cam)
		cam.SetDesiredResolution(camConfig.FrameWidth, camConfig.FrameHeight)
		actualWidth, actualHeight := cam.GetActualResolution()
		cam.Start(camConfig.FrameRate)

		fmt.Printf("========================================\n")
		fmt.Printf("Loaded and started camera: %s -> %d\n", camConfig.Name, camConfig.Device)
		fmt.Printf("----------------------------------------\n")
		fmt.Printf("Name: %s\nDevice: %d\nFramerate: %d\nFrame Width: %d\nFrame Height: %d\nActual Resolution: %.2f x %.2f\n",
			camConfig.Name, camConfig.Device, camConfig.FrameRate, camConfig.FrameWidth, camConfig.FrameHeight, actualWidth, actualHeight)
		fmt.Printf("----------------------------------------\n")
		fmt.Printf("Modes:\n")
		for camMode, mode := range camConfig.Modes {
			fmt.Printf(" \nMode: %s\n  Brightness: %.2f\n  Contrast: %.2f\n  Rotate: %d\n  Flip: %d\n  Saturation: %.2f\n  Quality: %d\n  Output Frame Width: %d\n  Output Frame Height: %d\n",
				camMode, mode.Brightness, mode.Contrast, mode.Rotate, mode.Flip, mode.Saturation, mode.Quality, mode.OutFrameWidth, mode.OutFrameHeight)
		}
		fmt.Printf("========================================\n")
	}
}

func Initialize() error {
	fmt.Println("initializing cameras")

	Config = config.Global
	Server = NewMultiCamServer()

	loadCameras()

	return nil
}
