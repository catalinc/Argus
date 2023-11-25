package argus

import (
	"encoding/json"
	"os"
	"time"
)

// MailConfig holds mail sender configuration
type MailConfig struct {
	From           string `json:"from"`
	To             string `json:"to"`
	ServerHost     string `json:"serverHost"`
	ServerPort     int    `json:"serverPort"`
	ServerUser     string `json:"serverUser"`
	ServerPassword string `json:"serverPassword"`
}

// Configuration holds application wide parameters
type Configuration struct {
	Fps         uint          `json:"fps"`
	DeviceID    string        `json:"deviceId"`
	MinInterval time.Duration `json:"minInterval"`
	MinArea     float64       `json:"minArea"`
	ShowVideo   bool          `json:"showVideo"`
	Handlers    []string      `json:"handlers"`
	DataDir     string        `json:"dataDir"`
	MailConfig  MailConfig    `json:"mailConfig"`
}

// LoadConfiguration loads configuration from file
func LoadConfiguration(path string) (Configuration, error) {
	var config Configuration
	file, err := os.Open(path)
	if err != nil {
		return config, err
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	if err := dec.Decode(&config); err != nil {
		return config, err
	}
	return config, nil
}

// DefaultConfiguration returns a configuration suitable for most cases
func DefaultConfiguration() Configuration {
	return Configuration{
		Fps:         10,
		DeviceID:    "0",
		MinInterval: time.Second * 5,
		MinArea:     10000,
		ShowVideo:   true,
		Handlers:    []string{"console", "archive"},
		DataDir:     "data"}
}
