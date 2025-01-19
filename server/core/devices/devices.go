// core/devices/devices.go
package devices

import (
	"fmt"

	"smuggr.xyz/gatecam/common/config"
)

var Config *config.GlobalConfig
var Server *DevicesServer

func loadDevices() {
	for _, devConfig := range Config.Devices {
		dev := NewDevice(devConfig)
		Server.AddDevice(dev)
		fmt.Printf("Loaded device: %s with IP: %s\n", devConfig.Name, devConfig.IP)
	}
}

func Initialize() {
	fmt.Println("Initializing devices")
	Config = &config.Global

	Server = NewDevicesServer()
	loadDevices()
}
