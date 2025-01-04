// core/cameras/camera.go
package cameras

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"sync"
	"time"

	"gocv.io/x/gocv"
	"smuggr.xyz/gate-cam/common/config"
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
	capture     *gocv.VideoCapture
	mu          sync.Mutex
	running     bool
	config      config.CameraConfig
	detections  []Entity
	detectionMu sync.Mutex
	net         gocv.Net
	outputs	    map[config.CameraMode]CameraModeOutput
}

func NewCamera(camConfig config.CameraConfig) (*Camera, error) {
	cap, err := gocv.OpenVideoCapture(camConfig.Device)
	if err != nil {
		return nil, fmt.Errorf("error opening camera %d: %v", camConfig.Device, err)
	}

	if camConfig.FrameWidth > 0 {
		cap.Set(gocv.VideoCaptureFrameWidth, float64(camConfig.FrameWidth))
	} else {
		return nil, fmt.Errorf("frame width must be greater than 0")
	}

	if camConfig.FrameHeight > 0 {
		cap.Set(gocv.VideoCaptureFrameHeight, float64(camConfig.FrameHeight))
	} else {
		return nil, fmt.Errorf("frame height must be greater than 0")
	}

	net := gocv.ReadNetFromCaffe(os.Getenv("MOBILENET_PROTOTXT"), os.Getenv("MOBILENET_MODEL"))
	if net.Empty() {
		return nil, fmt.Errorf("error loading MobileNet-SSD model")
	}

	outputs := map[config.CameraMode]CameraModeOutput{}

	for camMode, mode := range camConfig.Modes {
		if mode.OutFrameWidth == 0 || mode.OutFrameHeight == 0 {
			mode.OutFrameHeight = camConfig.FrameHeight
			mode.OutFrameWidth = camConfig.FrameWidth
		}
		outputs[camMode] = CameraModeOutput{
			config: mode,
		}
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

func matToRGB565(mat *gocv.Mat) ([]byte, error) {
    if mat.Type() != gocv.MatTypeCV8UC3 {
        return nil, fmt.Errorf("unexpected Mat type: %d", mat.Type())
    }

    rows := mat.Rows()
    cols := mat.Cols()

    rgb565 := make([]byte, rows*cols*2)

    for y := 0; y < rows; y++ {
        for x := 0; x < cols; x++ {
            b := mat.GetUCharAt(y, x*3+0)
            g := mat.GetUCharAt(y, x*3+1)
            r := mat.GetUCharAt(y, x*3+2)

            rgb565Value := uint16(((uint16(r) & 0xF8) << 8) | ((uint16(g) & 0xFC) << 3) | (uint16(b) >> 3))

            idx := (y*cols + x) * 2
            rgb565[idx] = byte(rgb565Value >> 8)
            rgb565[idx+1] = byte(rgb565Value & 0xFF)
        }
    }

    return rgb565, nil
}

func (cam *Camera) detectObjects(frame gocv.Mat) []Entity {
	detections := []Entity{}

	blob := gocv.BlobFromImage(frame, 1.0/127.5, image.Pt(300, 300), gocv.NewScalar(127.5, 127.5, 127.5, 0), false, false)
	defer blob.Close()

	cam.net.SetInput(blob, "")
	output := cam.net.Forward("")

	nrows := output.Total() / 7
	for i := 0; i < nrows; i++ {
		confidence := output.GetFloatAt(0, i*7+2)
		if confidence > 0.5 {
			classID := int(output.GetFloatAt(0, i*7+1))
			if classID == 7 || classID == 15 {
				x1 := int(output.GetFloatAt(0, i*7+3) * float32(frame.Cols()))
				y1 := int(output.GetFloatAt(0, i*7+4) * float32(frame.Rows()))
				x2 := int(output.GetFloatAt(0, i*7+5) * float32(frame.Cols()))
				y2 := int(output.GetFloatAt(0, i*7+6) * float32(frame.Rows()))
				detections = append(detections, Entity{
					Rect:       image.Rect(x1, y1, x2, y2),
					Confidence: confidence,
					Label:      map[int]string{7: "Car", 15: "Person"}[classID],
					Timestamp:  time.Now(),
				})
			}
		}
	}

	return detections
}

func (cam *Camera) drawDetections(frame *gocv.Mat, detections []Entity) {
	for _, det := range detections {
		gocv.Rectangle(frame, det.Rect, color.RGBA{0, 255, 0, 0}, 2)

		label := fmt.Sprintf(det.Label + " %f", det.Confidence)
		gocv.PutText(frame, label, image.Pt(det.Rect.Min.X, det.Rect.Min.Y-10),
			gocv.FontHersheySimplex, 1.0, color.RGBA{0, 255, 0, 0}, 2)
	}
}

func (cam *Camera) rotateImage(mat gocv.Mat, modeConfig config.CameraModeConfig) gocv.Mat {
	switch modeConfig.Rotate {
	case 90:
		gocv.Rotate(mat, &mat, gocv.Rotate90Clockwise)
	case 180:
		gocv.Rotate(mat, &mat, gocv.Rotate180Clockwise)
	case 270:
		gocv.Rotate(mat, &mat, gocv.Rotate90CounterClockwise)
	}
	return mat
}

func (cam *Camera) flipImage(mat gocv.Mat, modeConfig config.CameraModeConfig) gocv.Mat {
	if modeConfig.Flip == -1 || modeConfig.Flip == 0 || modeConfig.Flip == 1 {
		gocv.Flip(mat, &mat, modeConfig.Flip)
	}
	return mat
}

func (cam *Camera) adjustBrightnessContrast(mat gocv.Mat, modeConfig config.CameraModeConfig) gocv.Mat {
	alpha := modeConfig.Contrast
	beta := modeConfig.Brightness

	gocv.ConvertScaleAbs(mat, &mat, alpha, beta)
	return mat
}

func (cam *Camera) scaleImage(mat gocv.Mat, modeConfig config.CameraModeConfig) gocv.Mat {
    if modeConfig.OutFrameWidth == 0 || modeConfig.OutFrameHeight == 0 {
        return mat
    }

    if mat.Cols() == modeConfig.OutFrameWidth && mat.Rows() == modeConfig.OutFrameHeight {
		return mat.Clone()
	}

	resized := gocv.NewMat()
	gocv.Resize(mat, &resized, image.Pt(modeConfig.OutFrameWidth, modeConfig.OutFrameHeight), 0, 0, gocv.InterpolationLinear)
	return resized
}

func (cam *Camera) applyPostProcessing(mat gocv.Mat, modeConfig config.CameraModeConfig) gocv.Mat {
	mat = cam.rotateImage(mat, modeConfig)
	mat = cam.flipImage(mat, modeConfig)
	mat = cam.adjustBrightnessContrast(mat, modeConfig)

    mat = cam.scaleImage(mat, modeConfig)

	return mat
}

func (cam *Camera) grabFrameRGB565(mat gocv.Mat) ([]byte, error) {
    rgb565Data, err := matToRGB565(&mat)
    if err != nil {
        return nil, fmt.Errorf("failed to convert mat to RGB565: %v", err)
    }
	
	if len(rgb565Data) != 160*128*2 {
        return nil, fmt.Errorf("invalid frame size: expected %d, got %d", 160*128*2, len(rgb565Data)) 
    }

    return rgb565Data, nil
}

func (cam *Camera) grabFrameGrayscale(mat gocv.Mat) ([]byte, error) {
    grayMat := gocv.NewMat()
    defer grayMat.Close()
    gocv.CvtColor(mat, &grayMat, gocv.ColorBGRToGray)

    return grayMat.ToBytes(), nil
}

func (cam *Camera) grabFrameJPEG(mat gocv.Mat, modeConfig config.CameraModeConfig) ([]byte, error) {
	img, err := mat.ToImage()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: modeConfig.Quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (cam *Camera) checkAndRecoverCamera() {
    cam.mu.Lock()
    defer cam.mu.Unlock()

    if !cam.capture.IsOpened() {
        fmt.Printf("camera %s disconnected, attempting to reinitialize\n", cam.Name)
        cam.capture.Close()
        newCap, err := gocv.OpenVideoCapture(cam.Device)
        if err != nil {
            fmt.Printf("failed to reinitialize camera %s: %v\n", cam.Name, err)
            return
        }
        cam.capture = newCap
    }
}

func (cam *Camera) grabFrame(mode config.CameraMode) ([]byte, error) {
    cam.mu.Lock()
    defer cam.mu.Unlock()

    if cam.capture == nil || !cam.capture.IsOpened() {
        fmt.Printf("camera %s is not initialized or disconnected\n", cam.Name)
        cam.checkAndRecoverCamera()
        return nil, nil
    }

    mat := gocv.NewMat()
    defer mat.Close()

    if !cam.capture.Read(&mat) || mat.Empty() {
        fmt.Printf("camera %s failed to read frame\n", cam.Name)
        return nil, nil
    }

    modeConfig := cam.config.Modes[mode]
    mat = cam.applyPostProcessing(mat, modeConfig)

    detections := cam.detectObjects(mat)
    cam.detectionMu.Lock()
    cam.detections = detections
    cam.detectionMu.Unlock()

    cam.drawDetections(&mat, detections)

    switch mode {
    case config.ModeGrayscaleFrame:
        return cam.grabFrameGrayscale(mat)
    case config.ModeColorFrame:
        return cam.grabFrameRGB565(mat)
    case config.ModeJPEGStream:
        return cam.grabFrameJPEG(mat, modeConfig)
    }

    return nil, fmt.Errorf("unsupported camera mode: %s", mode)
}

func (cam *Camera) streamFrames(frameRate int) {
    interval := time.Duration(1000 / frameRate) * time.Millisecond
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
	if width > 0 {
		cam.capture.Set(gocv.VideoCaptureFrameWidth, float64(width))
	}
	if height > 0 {
		cam.capture.Set(gocv.VideoCaptureFrameHeight, float64(height))
	}
}

func (cam *Camera) GetActualResolution() (float64, float64) {
	return cam.capture.Get(gocv.VideoCaptureFrameWidth), cam.capture.Get(gocv.VideoCaptureFrameHeight)
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