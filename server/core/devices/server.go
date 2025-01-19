package devices

import (
    "strconv"
    "sync"
)

type DevicesServer struct {
    devices map[string]*Device
    mu      sync.RWMutex
}

func NewDevicesServer() *DevicesServer {
    return &DevicesServer{
        devices: make(map[string]*Device),
    }
}

func (ds *DevicesServer) AddDevice(device *Device) {
    ds.mu.Lock()
    defer ds.mu.Unlock()
    ds.devices[device.Name] = device
}

func (ds *DevicesServer) GetDevice(id string) (*Device, bool) {
    ds.mu.RLock()
    defer ds.mu.RUnlock()

    if devID, err := strconv.ParseUint(id, 10, 64); err == nil {
        for _, dev := range ds.devices {
            if uint64(dev.Order) == devID {
                return dev, true
            }
        }
    }
    dev, ok := ds.devices[id]
    return dev, ok
}