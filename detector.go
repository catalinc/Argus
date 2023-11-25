package argus

import (
	"errors"
	"image"
	"image/color"
	"time"

	"gocv.io/x/gocv"
)

// MotionEvent holds the image frame and the time when motion is detected
type MotionEvent struct {
	Frame     image.Image
	Timestamp time.Time
}

// MotionDetector uses the video capture device to detect motion
type MotionDetector interface {
	// OpenDevice opens video capture device for motion detection
	OpenDevice(deviceID string) error
	// DetectMotion motion in the video capture
	// Returns non nil MotionEvent when motion has been detected
	// Called at fixed intervals by the driver program
	DetectMotion(showVideo bool, minArea float64) (*MotionEvent, error)
	// Close video capture device and perform any additional cleanup
	Close()
}

// NewMotionDetector creates a motion detector based on OpenCV
func NewMotionDetector() MotionDetector {
	return &openCVMotionDetector{}
}

type openCVMotionDetector struct {
	webcam    *gocv.VideoCapture
	window    *gocv.Window
	img       gocv.Mat
	imgDelta  gocv.Mat
	imgThresh gocv.Mat
	mog2      gocv.BackgroundSubtractorMOG2
	kernel    gocv.Mat
}

func (d *openCVMotionDetector) OpenDevice(deviceID string) error {
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		return err
	}
	d.webcam = webcam
	d.window = gocv.NewWindow("Motion Detector")
	d.img = gocv.NewMat()
	d.imgDelta = gocv.NewMat()
	d.imgThresh = gocv.NewMat()
	d.mog2 = gocv.NewBackgroundSubtractorMOG2()
	d.kernel = gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	return nil
}

func (d *openCVMotionDetector) DetectMotion(showVideo bool, minimumArea float64) (*MotionEvent, error) {
	if ok := d.webcam.Read(&d.img); !ok {
		return nil, errors.New("video capture device is closed")
	}
	if d.img.Empty() {
		return nil, nil
	}

	motion := false

	// First phase of cleaning up image, obtain foreground only
	d.mog2.Apply(d.img, &d.imgDelta)

	// Remaining cleanup of the image to use for finding contours
	// First use threshold
	gocv.Threshold(d.imgDelta, &d.imgThresh, 25, 255, gocv.ThresholdBinary)

	// Then dilate
	gocv.Dilate(d.imgThresh, &d.imgThresh, d.kernel)

	// Now find contours
	contours := gocv.FindContours(d.imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)

		area := gocv.ContourArea(c)
		if area < minimumArea {
			continue
		}

		motion = true

		rect := gocv.BoundingRect(c)
		gocv.Rectangle(&d.img, rect, color.RGBA{255, 0, 0, 0}, 2)
	}

	var status string
	var statusColor color.RGBA
	if motion {
		status = "Motion detected"
		statusColor = color.RGBA{255, 0, 0, 0}
	} else {
		status = "Ready"
		statusColor = color.RGBA{0, 255, 0, 0}
	}
	gocv.PutText(&d.img, status, image.Pt(10, 20), gocv.FontHersheyPlain, 1.2, statusColor, 2)

	if showVideo {
		d.window.IMShow(d.img)
		d.window.WaitKey(1)
	}

	if motion {
		frame, err := d.img.ToImage()
		if err != nil {
			return nil, nil
		}
		return &MotionEvent{Frame: frame, Timestamp: time.Now()}, nil
	}
	return nil, nil
}

func (d *openCVMotionDetector) Close() {
	_ = d.webcam.Close()
	_ = d.window.Close()
	_ = d.img.Close()
	_ = d.imgDelta.Close()
	_ = d.imgThresh.Close()
	_ = d.mog2.Close()
	_ = d.kernel.Close()
}
