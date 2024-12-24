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


type Camera struct {
	Name       string
	Device     int
	capture    *gocv.VideoCapture
	mu         sync.Mutex
	running    bool
	lastFrame  []byte
	lastErr    error
	config     config.CameraConfig
	detections []Entity
	detectionMu sync.Mutex
	net        gocv.Net
}

func NewCamera(camConfig config.CameraConfig) (*Camera, error) {
	cap, err := gocv.OpenVideoCapture(camConfig.Device)
	if err != nil {
		return nil, fmt.Errorf("error opening camera %d: %v", camConfig.Device, err)
	}

	if camConfig.FrameWidth > 0 {
		cap.Set(gocv.VideoCaptureFrameWidth, float64(camConfig.FrameWidth))
	}
	if camConfig.FrameHeight > 0 {
		cap.Set(gocv.VideoCaptureFrameHeight, float64(camConfig.FrameHeight))
	}

	net := gocv.ReadNetFromCaffe(os.Getenv("MOBILENET_PROTOTXT"), os.Getenv("MOBILENET_MODEL"))
	if net.Empty() {
		return nil, fmt.Errorf("error loading MobileNet-SSD model")
	}

	return &Camera{
		Name:       camConfig.Name,
		Device:     camConfig.Device,
		capture:    cap,
		config:     camConfig,
		net:        net,
		detections: []Entity{},
	}, nil
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

func (cam *Camera) rotateImage(mat gocv.Mat) gocv.Mat {
	switch cam.config.Rotate {
	case 90:
		gocv.Rotate(mat, &mat, gocv.Rotate90Clockwise)
	case 180:
		gocv.Rotate(mat, &mat, gocv.Rotate180Clockwise)
	case 270:
		gocv.Rotate(mat, &mat, gocv.Rotate90CounterClockwise)
	}
	return mat
}

func (cam *Camera) flipImage(mat gocv.Mat) gocv.Mat {
	if cam.config.Flip == -1 || cam.config.Flip == 0 || cam.config.Flip == 1 {
		gocv.Flip(mat, &mat, cam.config.Flip)
	}
	return mat
}

func (cam *Camera) adjustBrightnessContrast(mat gocv.Mat) gocv.Mat {
	alpha := cam.config.Contrast
	beta := cam.config.Brightness

	gocv.ConvertScaleAbs(mat, &mat, alpha, beta)
	return mat
}

func (cam *Camera) scaleImage(mat gocv.Mat) gocv.Mat {
    if cam.config.OutFrameWidth == 0 || cam.config.OutFrameHeight == 0 {
        return mat
    }

    if mat.Cols() == cam.config.OutFrameWidth && mat.Rows() == cam.config.OutFrameHeight {
		return mat.Clone()
	}

	resized := gocv.NewMat()
	gocv.Resize(mat, &resized, image.Pt(cam.config.OutFrameWidth, cam.config.OutFrameHeight), 0, 0, gocv.InterpolationLinear)
	return resized
}

func (cam *Camera) applyPostProcessing(mat gocv.Mat) gocv.Mat {
	mat = cam.rotateImage(mat)
	mat = cam.flipImage(mat)
	mat = cam.adjustBrightnessContrast(mat)

    mat = cam.scaleImage(mat)

	return mat
}

func (cam *Camera) grabFrame() ([]byte, error) {
	if cam.capture == nil {
		return nil, fmt.Errorf("camera %s not initialized", cam.Name)
	}

	mat := gocv.NewMat()
	defer mat.Close()

	if ok := cam.capture.Read(&mat); !ok {
		return nil, fmt.Errorf("cannot read frame from camera %s", cam.Name)
	}
	if mat.Empty() {
		return nil, fmt.Errorf("empty frame from camera %s", cam.Name)
	}

	mat = cam.applyPostProcessing(mat)

	detections := cam.detectObjects(mat)
	cam.detectionMu.Lock()
	cam.detections = detections
	cam.detectionMu.Unlock()

	cam.drawDetections(&mat, detections)

	img, err := mat.ToImage()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: cam.config.Quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (cam *Camera) streamFrames(frameRate int) {
	interval := time.Duration(1000/frameRate) * time.Millisecond
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
		frame, err := cam.grabFrame()
		cam.mu.Lock()
		if err != nil {
			cam.lastErr = err
		} else {
			cam.lastFrame = frame
			cam.lastErr = nil
		}
		cam.mu.Unlock()
	}
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

func (cam *Camera) SetDesiredResolution(width, height int) {
	if width > 0 {
		cam.capture.Set(gocv.VideoCaptureFrameWidth, float64(width))
	}
	if height > 0 {
		cam.capture.Set(gocv.VideoCaptureFrameHeight, float64(height))
	}
}

func (cam *Camera) ReadFrame() ([]byte, error) {
	cam.mu.Lock()
	defer cam.mu.Unlock()

	if cam.lastErr != nil {
		return nil, fmt.Errorf("camera %s error: %v", cam.Name, cam.lastErr)
	}
	if len(cam.lastFrame) == 0 {
		return nil, fmt.Errorf("camera %s has no frame yet", cam.Name)
	}

	out := make([]byte, len(cam.lastFrame))
	copy(out, cam.lastFrame)
	return out, nil
}

func (cam *Camera) Stop() {
	cam.mu.Lock()
	defer cam.mu.Unlock()
	cam.running = false
	if cam.capture != nil {
		cam.capture.Close()
		cam.capture = nil
	}
}
