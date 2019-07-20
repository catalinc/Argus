package main

import (
	"flag"
	"github.com/catalinc/argus"
	"log"
	"os"
	"os/signal"
	"time"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "configFile", "", "configuration file")
}

func main() {
	flag.Parse()

	var config argus.Configuration
	var err error
	if configFile != "" {
		if config, err = argus.LoadConfiguration(configFile); err != nil {
			log.Printf("Cannot load configuration from %s: %v\n", configFile, err)
			os.Exit(1)
		}
	} else {
		config = argus.DefaultConfiguration()
	}
	log.Printf("Starting...\n")

	runner := argus.NewRunner(config, argus.NewMotionDetector())
	if err := runner.Init(); err != nil {
		log.Printf("Initialization error: %v\n", err)
		os.Exit(1)
	}
	defer runner.Close()

	quit := make(chan bool)
	trap := make(chan os.Signal, 1)

	ticker := time.NewTicker(time.Duration(1000/config.Fps) * time.Millisecond)

	signal.Notify(trap, os.Interrupt)
	go func() {
		select {
		case sig := <-trap:
			log.Printf("Aborting: got %v\n", sig)
			quit <- true
		}
	}()

	for {
		select {
		case <-ticker.C:
			if err := runner.Run(); err != nil {
				log.Printf("Run error: %v\n", err)
				continue

			}
		case <-quit:
			log.Println("Bye")
			ticker.Stop()
			return
		}
	}
}
