// core/cameras/helpers.go
package cameras

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"time"

	"gocv.io/x/gocv"
	"smuggr.xyz/gate-cam/common/config"
)

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

func (cam *Camera) rotateImage(mat *gocv.Mat, modeConfig config.CameraModeConfig) {
	switch modeConfig.Rotate {
	case 90:
		gocv.Rotate(*mat, mat, gocv.Rotate90Clockwise)
	case 180:
		gocv.Rotate(*mat, mat, gocv.Rotate180Clockwise)
	case 270:
		gocv.Rotate(*mat, mat, gocv.Rotate90CounterClockwise)
	}
}

func (cam *Camera) flipImage(mat *gocv.Mat, modeConfig config.CameraModeConfig) {
    if modeConfig.Flip == -1 || modeConfig.Flip == 0 || modeConfig.Flip == 1 {
        gocv.Flip(*mat, mat, modeConfig.Flip)
    }
}

func (cam *Camera) adjustBrightnessContrast(mat *gocv.Mat, modeConfig config.CameraModeConfig) {
    alpha := modeConfig.Contrast
    beta := modeConfig.Brightness
    gocv.ConvertScaleAbs(*mat, mat, alpha, beta)
}

func (cam *Camera) scaleImage(mat *gocv.Mat, modeConfig config.CameraModeConfig) {
    if modeConfig.OutFrameWidth == 0 || modeConfig.OutFrameHeight == 0 {
        return
    }

    if mat.Cols() == modeConfig.OutFrameWidth && mat.Rows() == modeConfig.OutFrameHeight {
        return
    }

    resized := gocv.NewMat()
    gocv.Resize(*mat, &resized, image.Pt(modeConfig.OutFrameWidth, modeConfig.OutFrameHeight), 0, 0, gocv.InterpolationLinear)
    
    mat.Close()
    *mat = resized
}

func (cam *Camera) applyPostProcessing(mat *gocv.Mat, modeConfig config.CameraModeConfig) {
    cam.rotateImage(mat, modeConfig)
    cam.flipImage(mat, modeConfig)
    cam.adjustBrightnessContrast(mat, modeConfig)
    cam.scaleImage(mat, modeConfig)
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