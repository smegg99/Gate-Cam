// core/cameras/cameras.go
package cameras

import (
    "fmt"

    "smuggr.xyz/gate-cam/common/config"
)

var Config config.CamerasConfig
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
		cam.Start(camConfig.FrameRate)

		fmt.Printf("loaded and started camera: %s -> %d\n", camConfig.Name, camConfig.Device)
	}
}

func Initialize() error {
	fmt.Println("initializing cameras")

    Config = config.Global.Cameras
	Server = NewMultiCamServer()

	loadCameras()

	return nil
}