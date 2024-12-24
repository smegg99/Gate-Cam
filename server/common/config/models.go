// config/models.go
package config

type APIConfig struct {
	Port int16 `mapstructure:"port"`
}

type CameraConfig struct {
	Name           string  `mapstructure:"name"`
	Device         int     `mapstructure:"device"`
	FrameRate      int     `mapstructure:"frame_rate"`
	FrameWidth     int     `mapstructure:"frame_width"`
	FrameHeight    int     `mapstructure:"frame_height"`
	Brightness     float64 `mapstructure:"brightness"`
    Contrast       float64 `mapstructure:"contrast"`
    Rotate         int     `mapstructure:"rotate"`           // 0, 90, 180, 270
    Flip           int     `mapstructure:"flip"`             // -1=both axes, 0=x-axis, 1=y-axis
    Saturation     float64 `mapstructure:"saturation"`
	Quality	       int     `mapstructure:"quality"`          // jpeg quality
	OutFrameWidth  int     `mapstructure:"out_frame_width"`
	OutFrameHeight int     `mapstructure:"out_frame_height"`
}

type CamerasConfig struct {
	Cameras []CameraConfig `mapstructure:"cameras"`
}

type GlobalConfig struct {
	API     APIConfig     `mapstructure:"api"`
	Cameras CamerasConfig `mapstructure:"cameras"`
}