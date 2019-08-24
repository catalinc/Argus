package argus

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

const (
	filePattern = "2006-01-02-15-04-05"
	fileExt     = ".png"
	dateFormat  = "2006-01-02 15:04:05"
)

// MotionHandler represents a callback invoked when motion is detected
type MotionHandler interface {
	// Handle motion event
	Handle(event *MotionEvent) error
}

func newMotionHandler(handler string, config Configuration) (MotionHandler, error) {
	switch handler {
	case "console":
		return &consoleHandler{}, nil
	case "archive":
		return &archiveHandler{dataDir: config.DataDir}, nil
	case "mail":
		return &mailHandler{
			sender:  newMailSender(config),
			from:    config.MailConfig.From,
			to:      []string{config.MailConfig.To},
			dataDir: config.DataDir}, nil
	default:
		return nil, fmt.Errorf("unknown handler: %v", handler)
	}
}

type consoleHandler struct{}

func (r *consoleHandler) Handle(event *MotionEvent) error {
	log.Printf("Motion detected at %s\n", event.Timestamp.Format(dateFormat))
	return nil
}

type archiveHandler struct {
	dataDir string
}

func pathForImage(event *MotionEvent, dataDir string) string {
	filename := event.Timestamp.Format(filePattern) + fileExt
	return filepath.Join(dataDir, filename)
}

func (a *archiveHandler) Handle(event *MotionEvent) error {
	if _, err := os.Stat(a.dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(a.dataDir, os.ModePerm); err != nil {
			return err
		}
	}
	path := pathForImage(event, a.dataDir)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := png.Encode(file, event.Frame); err != nil {
		return err
	}
	log.Printf("Motion capture saved to %s\n", path)
	return nil
}

type mailHandler struct {
	sender  mailSender
	from    string
	to      []string
	dataDir string
}

func (m *mailHandler) Handle(event *MotionEvent) error {
	path := pathForImage(event, m.dataDir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	body := "Motion detected at " + event.Timestamp.Format(dateFormat)
	msg := mailMessage{
		from:       m.from,
		to:         m.to,
		subject:    "Motion detected",
		body:       body,
		attachment: path}
	err := m.sender.send(msg)
	log.Printf("Email sent with motion capture\n")
	return err
}
