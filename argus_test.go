package argus

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestRunner_Init(t *testing.T) {
	config := newMockConfiguration()
	detector := newMockMotionDetector()
	runner := NewRunner(config, detector)
	if err := runner.Init(); err != nil {
		t.Fatalf("initialization error: %v", err)
	}
	if cnt := detector.getCallCount("OpenDevice"); cnt != 1 {
		t.Fatalf("OpenDevice called %d times: expected one call", cnt)
	}

	if cnt := len(runner.handlers); cnt != 1 {
		t.Fatalf("got %d handlers registered: expected one handler", cnt)
	}
}

func TestRunner_Run(t *testing.T) {
	config := newMockConfiguration()
	detector := newMockMotionDetector()
	runner := NewRunner(config, detector)
	if err := runner.Init(); err != nil {
		t.Fatalf("initialization error: %v", err)
	}
	ops := 3
	for i := 0; i < ops; i++ {
		if err := runner.Run(); err != nil {
			t.Fatalf("run error: %v", err)
		}
	}
	if cnt := detector.getCallCount("DetectMotion"); cnt != ops {
		t.Fatalf("DetectMotion called %d times: expected %d calls", cnt, ops)
	}
	if cnt := detector.eventCount; cnt != ops {
		t.Fatalf("got %d motion events created: expected %d", cnt, ops)
	}
	if cnt := runner.eventCount; cnt != 1 {
		t.Fatalf("got %d motion events handled: expected one", cnt)
	}
}

func TestRunner_Close(t *testing.T) {
	config := newMockConfiguration()
	detector := newMockMotionDetector()
	runner := NewRunner(config, detector)
	if err := runner.Init(); err != nil {
		t.Fatalf("initialization error: %v", err)
	}
	runner.Close()
	if cnt := detector.getCallCount("Close"); cnt != 1 {
		t.Fatalf("Close called %d times: expected one call", cnt)
	}
}

func TestNewMotionHandler(t *testing.T) {
	config := newMockConfiguration()
	for _, typ := range []string{"reporter", "archive", "mail"} {
		_, err := newMotionHandler(typ, config)
		if err != nil {
			t.Fatalf("cannot create motion handler for type %v", typ)
		}
	}
	unsupported := "unsupported"
	_, err := newMotionHandler(unsupported, config)
	if err == nil {
		t.Fatalf("should not create motion handler for type %v", unsupported)
	}
}

func TestReporterHandler_Handle(t *testing.T) {
	handler := reporterHandler{}
	if err := handler.Handle(newTestMotionEvent()); err != nil {
		t.Fatalf("error while handling event: %v", err)
	}
}

func TestArchiveHandler_Handle(t *testing.T) {
	config := newMockConfiguration()
	handler := archiveHandler{dataDir: config.DataDir}
	event := newTestMotionEvent()
	if err := handler.Handle(event); err != nil {
		t.Fatalf("error while handling event: %v", err)
	}
	name := event.Timestamp.Format(filePattern) + fileExt
	path := filepath.Join(config.DataDir, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("%s image file does not exist", path)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("cannot delete image file %s", path)
	}
}

func TestMailHandler_Handle(t *testing.T) {
	config := newMockConfiguration()
	sender := newMockMailSender()
	handler := mailHandler{
		sender:  sender,
		from:    config.MailConfig.From,
		to:      []string{config.MailConfig.To},
		dataDir: config.DataDir}
	event := newTestMotionEvent()
	// Save image
	archive := archiveHandler{dataDir: config.DataDir}
	if err := archive.Handle(event); err != nil {
		t.Fatalf("error saving frame to file: %v", err)
	}
	name := event.Timestamp.Format(filePattern) + fileExt
	path := filepath.Join(config.DataDir, name)
	defer os.Remove(path)
	// Now actually handle the event
	if err := handler.Handle(event); err != nil {
		t.Fatalf("error while handling event: %v", err)
	}
	cnt := sender.getCallCount("send")
	if cnt != 1 {
		t.Fatalf("send called %d times: expected once", cnt)
	}
	expectedMsg := mailMessage{
		from: config.MailConfig.From,
		to:   []string{config.MailConfig.To}}
	expectedMsg.subject = "Motion detected"
	expectedMsg.body = fmt.Sprintf("Motion detected at %s", event.Timestamp.Format(dateFormat))
	expectedMsg.attachment = pathForImage(event, config.DataDir)
	actualMsg := sender.lastSentMessage
	if !reflect.DeepEqual(expectedMsg, actualMsg) {
		t.Fatalf("mail messages do not match: expected %v got %v", expectedMsg, actualMsg)
	}
	newEvent := newTestMotionEvent()
	newEvent.Timestamp = time.Now().AddDate(0, -1, -1)
	err := handler.Handle(newEvent)
	if err == nil {
		t.Fatalf("expecte error if the image file does not exists")
	}
}

func TestLoadConfiguration(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "tmp")
	if err != nil {
		t.Fatalf("cannot create temporary file")
	}
	defer os.Remove(tmpFile.Name())

	expectedCfg := Configuration{
		DeviceId:    "0",
		MinInterval: time.Second * 5,
		MinArea:     10000,
		ShowVideo:   false,
		Handlers:    []string{"reporter,archive"},
		DataDir:     "data",
		MailConfig: MailConfig{
			From:           "user@example.net",
			To:             "someone@example.net",
			ServerHost:     "smtp.example.net",
			ServerPort:     587,
			ServerUser:     "user@example.net",
			ServerPassword: "secret"},
	}

	enc := json.NewEncoder(tmpFile)
	if err := enc.Encode(expectedCfg); err != nil {
		t.Fatalf("cannot save configuration to file: %v", err)
	}

	actualCfg, err := LoadConfiguration(tmpFile.Name())
	if err != nil {
		t.Fatalf("cannot load saved configuration %v", err)
	}
	if !reflect.DeepEqual(expectedCfg, actualCfg) {
		t.Fatalf("configurations do not match: expected %v got %v", expectedCfg, actualCfg)
	}
}

func TestDefaultConfiguration(t *testing.T) {
	expectedCfg := Configuration{
		Fps:         10,
		DeviceId:    "0",
		MinInterval: time.Second * 5,
		MinArea:     10000,
		ShowVideo:   true,
		Handlers:    []string{"reporter", "archive"},
		DataDir:     "data"}
	defaultCfg := DefaultConfiguration()
	if !reflect.DeepEqual(expectedCfg, defaultCfg) {
		t.Fatalf("default configuration doesnt match: expected %v got %v", expectedCfg, defaultCfg)
	}
}

func newMockConfiguration() Configuration {
	return Configuration{
		DeviceId:    "0",
		MinInterval: time.Duration(30) * time.Second,
		MinArea:     10000,
		ShowVideo:   false,
		Handlers:    []string{"reporter"},
		DataDir:     os.TempDir(),
		MailConfig: MailConfig{
			From:           "user@example.net",
			To:             "someone@example.net",
			ServerHost:     "smtp.example.net",
			ServerPort:     465,
			ServerUser:     "user@example.net",
			ServerPassword: "secret"},
	}
}

func newRecorder() recorder {
	return recorder{calls: make(map[string]int)}
}

type recorder struct {
	calls map[string]int
}

func (r *recorder) getCallCount(name string) int {
	cnt, set := r.calls[name]
	if !set {
		return 0
	}
	return cnt
}

func (r *recorder) registerCall(name string) {
	r.calls[name] = r.getCallCount(name) + 1
}

func newMockMotionDetector() *mockMotionDetector {
	return &mockMotionDetector{recorder: newRecorder()}
}

type mockMotionDetector struct {
	recorder
	eventCount int
}

func (d *mockMotionDetector) OpenDevice(deviceId string) error {
	d.registerCall("OpenDevice")
	return nil
}

func (d *mockMotionDetector) DetectMotion(showVideo bool, minArea float64) (*MotionEvent, error) {
	d.registerCall("DetectMotion")
	d.eventCount++
	return newTestMotionEvent(), nil
}

func (d *mockMotionDetector) Close() {
	d.registerCall("Close")
}

func newMockMailSender() *mockMailSender {
	return &mockMailSender{recorder: newRecorder()}
}

type mockMailSender struct {
	recorder
	lastSentMessage mailMessage
}

func (m *mockMailSender) send(msg mailMessage) error {
	m.lastSentMessage = msg
	m.registerCall("send")
	return nil
}

func newTestMotionEvent() *MotionEvent {
	frame := image.NewRGBA(image.Rect(0, 0, 1, 1))
	return &MotionEvent{Timestamp: time.Now(), Frame: frame}
}
