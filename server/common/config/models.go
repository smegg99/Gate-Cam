// config/models.go
package config

type APIConfig struct {
	Port         int16 `mapstructure:"port"`
	ExternalPort int16 `mapstructure:"external_port"`
}

type CameraMode string

const (
	ModeGrayscaleFrame CameraMode = "grayscale_frame"
	ModeColorFrame     CameraMode = "color_frame"
	ModeJPEGStream     CameraMode = "jpeg_stream"
)

func (cm CameraMode) String() string {
	return string(cm)
}

type CameraModeConfig struct {
	Brightness     float64 `mapstructure:"brightness"`
	Contrast       float64 `mapstructure:"contrast"`
	Rotate         int     `mapstructure:"rotate"` // 0, 90, 180, 270
	Flip           int     `mapstructure:"flip"`   // -1=both axes, 0=x-axis, 1=y-axis
	Saturation     float64 `mapstructure:"saturation"`
	Quality        int     `mapstructure:"quality"`         // jpeg quality
	OutFrameWidth  int     `mapstructure:"out_frame_width"` // Any value > 0 can be used
	OutFrameHeight int     `mapstructure:"out_frame_height"`
}

type CameraConfig struct {
	Name         string                          `mapstructure:"name"`
	Device       int                             `mapstructure:"device"`
	Order        uint                            `mapstructure:"order"`
	AccessKeyEnv string                          `mapstructure:"access_key_env"` // Used to authenticate camera streams, currently only JPEG
	FrameRate    int                             `mapstructure:"frame_rate"`
	FrameWidth   int                             `mapstructure:"frame_width"` // Only values supported by the camera will be used
	FrameHeight  int                             `mapstructure:"frame_height"`
	Modes        map[CameraMode]CameraModeConfig `mapstructure:"modes"`
	IsDisplay    bool                            `mapstructure:"is_display"` // Streaming desktop doesn't work yet
	DisplayIndex int                             `mapstructure:"display_index"`
}

type DeviceConfig struct {
	Name         string `mapstructure:"name"`
	Order        uint   `mapstructure:"order"`
	IP           string `mapstructure:"ip"`
	Port         int    `mapstructure:"port"`
	AccessKeyEnv string `mapstructure:"access_key_env"`
}

type GlobalConfig struct {
	API     APIConfig      `mapstructure:"api"`
	Cameras []CameraConfig `mapstructure:"cameras"`
	Devices []DeviceConfig `mapstructure:"devices"`
}
