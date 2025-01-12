// core/cameras/camera.go
package cameras

import (
	"fmt"
	"image"
	"os"
	"sync"
	"time"

    "smuggr.xyz/gate-cam/common/config"

	"gocv.io/x/gocv"
    "github.com/kbinani/screenshot"
)

type Entity struct {
	Rect       image.Rectangle
	Confidence float32
	Label      string
	Timestamp  time.Time
}

type CameraModeOutput struct {
	lastFrame   []byte
	lastErr     error
	config      config.CameraModeConfig
}

type Camera struct {
	Name        string
	Device      int
	Order       uint
	capture     *gocv.VideoCapture
	mu          sync.Mutex
	running     bool
	config      config.CameraConfig
	detections  []Entity
	// detectionMu sync.Mutex
	net         gocv.Net
	outputs	    map[config.CameraMode]CameraModeOutput
}

func NewCamera(camConfig config.CameraConfig) (*Camera, error) {
    var cap *gocv.VideoCapture
    var err error

    // Skip opening the physical camera if IsDisplay is true
    if !camConfig.IsDisplay {
        cap, err = gocv.OpenVideoCapture(camConfig.Device)
        if err != nil {
            return nil, fmt.Errorf("error opening camera %d: %v", camConfig.Device, err)
        }

        if camConfig.FrameWidth > 0 {
            cap.Set(gocv.VideoCaptureFrameWidth, float64(camConfig.FrameWidth))
        }
        if camConfig.FrameHeight > 0 {
            cap.Set(gocv.VideoCaptureFrameHeight, float64(camConfig.FrameHeight))
        }
    } else {
        // If capturing the screen, we do nothing here.
        // The capture device remains nil. That’s intentional.
    }

    net := gocv.ReadNetFromCaffe(os.Getenv("MOBILENET_PROTOTXT"), os.Getenv("MOBILENET_MODEL"))
    if net.Empty() {
        return nil, fmt.Errorf("error loading MobileNet-SSD model")
    }

    outputs := make(map[config.CameraMode]CameraModeOutput)
    for camMode, mode := range camConfig.Modes {
        if mode.OutFrameWidth == 0 {
            mode.OutFrameWidth = camConfig.FrameWidth
        }
        if mode.OutFrameHeight == 0 {
            mode.OutFrameHeight = camConfig.FrameHeight
        }
        outputs[camMode] = CameraModeOutput{ config: mode }
    }

    return &Camera{
        Name:       camConfig.Name,
        Device:     camConfig.Device,
        capture:    cap,
        config:     camConfig,
        net:        net,
        detections: []Entity{},
        outputs:    outputs,
    }, nil
}

func (cam *Camera) GetAccessKey() string {
	return os.Getenv(cam.config.AccessKeyEnv)
}

func (cam *Camera) grabScreenMat() (gocv.Mat, error) {
	displayIndex := cam.config.DisplayIndex
	bounds := screenshot.GetDisplayBounds(displayIndex)
	if bounds.Empty() {
		return gocv.NewMat(), fmt.Errorf("invalid display bounds for index %d", displayIndex)
	}

	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return gocv.NewMat(), fmt.Errorf("failed to capture display %d: %v", displayIndex, err)
	}

	matRGB, err := gocv.ImageToMatRGB(img)
	if err != nil { 
		return gocv.NewMat(), fmt.Errorf("failed to convert screenshot to Mat: %v", err)
	}
	defer matRGB.Close()

	matBGR := gocv.NewMat()
	gocv.CvtColor(matRGB, &matBGR, gocv.ColorRGBToBGR)

	return matBGR, nil
}

func (cam *Camera) grabFrame(mode config.CameraMode) ([]byte, error) {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	var mat gocv.Mat
	var err error

	if cam.config.IsDisplay {
		mat, err = cam.grabScreenMat()
		if err != nil {
			fmt.Printf("camera %s failed to capture screen: %v\n", cam.Name, err)
			return nil, nil
		}
	} else {
		if cam.capture == nil || !cam.capture.IsOpened() {
			fmt.Printf("camera %s is not initialized or disconnected\n", cam.Name)
			cam.checkAndRecoverCamera()
			return nil, nil
		}

		mat = gocv.NewMat()
		if !cam.capture.Read(&mat) || mat.Empty() {
			fmt.Printf("camera %s failed to read frame\n", cam.Name)
			return nil, nil
		}
	}
	defer mat.Close()

	modeConfig := cam.config.Modes[mode]
	mat = cam.applyPostProcessing(mat, modeConfig)

	// TODO: Fix detections causing a memory leak
	// detections := cam.detectObjects(mat)
	// cam.detectionMu.Lock()
	// cam.detections = detections
	// cam.detectionMu.Unlock()

	// cam.drawDetections(&mat, detections)

	switch mode {
	case config.ModeGrayscaleFrame:
		return cam.grabFrameGrayscale(mat)
	case config.ModeColorFrame:
		return cam.grabFrameRGB565(mat)
	case config.ModeJPEGStream:
		return cam.grabFrameJPEG(mat, modeConfig)
	default:
		return nil, fmt.Errorf("unsupported camera mode: %s", mode)
	}
}

func (cam *Camera) streamFrames(frameRate int) {
	interval := time.Duration(1000/frameRate) * time.Millisecond
	var wg sync.WaitGroup

	for mode, output := range cam.outputs {
		wg.Add(1)
		go func(mode config.CameraMode, output CameraModeOutput) {
			defer wg.Done()
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				cam.mu.Lock()
				if !cam.running {
					cam.mu.Unlock()
					break
				}
				cam.mu.Unlock()

				<-ticker.C
				frame, err := cam.grabFrame(mode)

				cam.mu.Lock()
				if err != nil {
					cam.outputs[mode] = CameraModeOutput{
						lastFrame: nil,
						lastErr:   err,
						config:    output.config,
					}
				} else {
					cam.outputs[mode] = CameraModeOutput{
						lastFrame: frame,
						lastErr:   nil,
						config:    output.config,
					}
				}
				cam.mu.Unlock()
			}
		}(mode, output)
	}

	wg.Wait()
}

func (cam *Camera) SetDesiredResolution(width, height int) {
	if cam.config.IsDisplay {
		cam.config.FrameWidth = width
		cam.config.FrameHeight = height
		fmt.Printf("SetDesiredResolution() called on display cam %s. Updated config only.\n", cam.Name)
		return
	}

	if cam.capture != nil && cam.capture.IsOpened() {
		if width > 0 {
			cam.capture.Set(gocv.VideoCaptureFrameWidth, float64(width))
		}
		if height > 0 {
			cam.capture.Set(gocv.VideoCaptureFrameHeight, float64(height))
		}
	}
}

func (cam *Camera) GetActualResolution() (float64, float64) {
	if cam.config.IsDisplay {
		bounds := screenshot.GetDisplayBounds(cam.config.DisplayIndex)
		if bounds.Empty() {
			return float64(cam.config.FrameWidth), float64(cam.config.FrameHeight)
		}
		return float64(bounds.Dx()), float64(bounds.Dy())
	}

	if cam.capture != nil && cam.capture.IsOpened() {
		w := cam.capture.Get(gocv.VideoCaptureFrameWidth)
		h := cam.capture.Get(gocv.VideoCaptureFrameHeight)
		return w, h
	}

	return 0, 0
}

func (cam *Camera) ReadFrame(mode config.CameraMode) ([]byte, error) {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	output := cam.outputs[mode]
	if output.lastErr != nil {
		return nil, fmt.Errorf("camera %s error: %v", cam.Name, output.lastErr)
	}

	out := make([]byte, len(output.lastFrame))
	copy(out, output.lastFrame)
	return out, nil
}

func (cam *Camera) Start(frameRate int) {
	cam.mu.Lock()
	if cam.running {
		cam.mu.Unlock()
		return
	}
	cam.running = true
	cam.mu.Unlock()

	go cam.streamFrames(frameRate)
}

func (cam *Camera) Stop() {
	cam.mu.Lock()
	cam.running = false
	cam.mu.Unlock()

	if cam.capture != nil {
		cam.capture.Close()
		cam.capture = nil
	}
}