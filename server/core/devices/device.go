// devices/device.go
package devices

import (
	"os"

	"smuggr.xyz/gatecam/common/config"
)

type Device struct {
	Name   string
	Order  uint
	config *config.DeviceConfig
}

func NewDevice(devConfig config.DeviceConfig) *Device {
	return &Device{
		Name:   devConfig.Name,
		Order:  devConfig.Order,
		config: &devConfig,
	}
}

func (d *Device) GetAccessKey() string {
	return os.Getenv(d.config.AccessKeyEnv)
}

func (d *Device) GetIP() string {
	return d.config.IP
}

func (d *Device) GetPort() int {
	return d.config.Port
}