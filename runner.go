package argus

import (
	"log"
	"time"
)

// Runner is the main driver for motion detection and handling
type Runner struct {
	config     Configuration
	detector   MotionDetector
	handlers   []MotionHandler
	threshold  time.Time // Used to cap detection handling at a specified interval
	eventCount int       // The number of handled motion events
}

// NewRunner returns a new Runner instance
func NewRunner(config Configuration, detector MotionDetector) *Runner {
	return &Runner{config: config, detector: detector}
}

// Init opens the video capture device and registers motion handlers
func (r *Runner) Init() error {
	err := r.detector.OpenDevice(r.config.DeviceId)
	if err != nil {
		return err
	}
	for _, name := range r.config.Handlers {
		h, err := newMotionHandler(name, r.config)
		if err != nil {
			return err
		}
		log.Printf("Adding motion handler: %s\n", name)
		r.handlers = append(r.handlers, h)
	}
	r.threshold = time.Now()
	return nil
}

// Run contains the logic to detect and handle motion
// Called periodically by the main program loop
func (r *Runner) Run() error {
	evt, err := r.detector.DetectMotion(r.config.ShowVideo, r.config.MinArea)
	if err != nil {
		return err
	}
	if evt != nil {
		if evt.Timestamp.After(r.threshold) {
			r.eventCount++
			r.threshold = evt.Timestamp.Add(r.config.MinInterval)
			go func() {
				for _, h := range r.handlers {
					if err := h.Handle(evt); err != nil {
						log.Printf("Handler error: %v\n", err)
					}
				}
				log.Printf("%d motion events handled\n", r.eventCount)
			}()
		}
	}
	return nil
}

// Close closes the video capture device
func (r *Runner) Close() {
	r.detector.Close()
}
