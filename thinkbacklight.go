package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"
)

// Config file structure type
type Config struct {
	BrightnessDevPath  string `yaml:"brightnessDevPath"`
	PowerSupplyDevPath string `yaml:"powerSupplyDevPath"`
	State              struct {
		OnBat     string `yaml:"BrightnessOnBat"`
		OnAC      string `yaml:"BrightnessOnAC"`
		PrevState string `yaml:"PrevState"`
	} `yaml:"State"`
}

// checkProcessError function checks err variable, terminates execution and returns corresponding exitcode
func checkProcessError(err error, exitcode int) {
	if err != nil {
		fmt.Println(err)
		log.Println("exitcode: " + string(exitcode))
		os.Exit(exitcode)
	}
}

func configFile(cfg *Config, configName string, action string) {
	switch action {
	case "read":
		{
			f, err := os.Open(configName)
			checkProcessError(err, 11)
			defer f.Close()
			decoder := yaml.NewDecoder(f)
			err = decoder.Decode(cfg)
			checkProcessError(err, 12)
		}
	case "write":
		{
			f, err := os.Create(configName)
			defer f.Close()
			checkProcessError(err, 13)
			encoder := yaml.NewEncoder(f)
			err = encoder.Encode(cfg)
			checkProcessError(err, 14)
			err = encoder.Close()
			checkProcessError(err, 15)

		}
	}
}

func getBackliteStatus(backlitePath string) string {
	f, err := ioutil.ReadFile(backlitePath)
	checkProcessError(err, 6)
	return (strings.TrimSuffix(string(f), "\n"))
}

func getPowerStatus(powerPath string) string {
	f, err := ioutil.ReadFile(powerPath)
	checkProcessError(err, 7)
	return (strings.TrimSuffix(string(f), "\n"))
}

func setBrightness(cfg *Config, level string, errorCode int) {
	d1 := []byte("")
	log.Println("Setting brightness level: " + level)
	d1 = []byte(level)
	err := ioutil.WriteFile(cfg.BrightnessDevPath, d1, 0644)
	checkProcessError(err, errorCode)
}

// Write a pid file, but first make sure it doesn't exist with a running pid.
func writePidFile(pidFile string) error {
	// Read in the pid file as a slice of bytes.
	if piddata, err := ioutil.ReadFile(pidFile); err == nil {
		// Convert the file contents to an integer.
		if pid, err := strconv.Atoi(string(piddata)); err == nil {
			// Look for the pid in the process list.
			if process, err := os.FindProcess(pid); err == nil {
				// Send the process a signal zero kill.
				if err := process.Signal(syscall.Signal(0)); err == nil {
					// We only get an error if the pid isn't running, or it's not ours.
					return fmt.Errorf("pid already running: %d", pid)
				}
			}
		}
	}
	// If we get here, then the pidfile didn't exist,
	// or the pid in it doesn't belong to the user running this app.
	return ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
}

func main() {

	err := writePidFile("/run/thinkbacklight.pid")
	checkProcessError(err, 131)
	configStr := flag.String("config", "config.yaml", "Path to a config file")
	flag.Parse()
	log.Println("Using config file " + *configStr)

	var cfg Config
	configFile(&cfg, *configStr, "read")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		for sig := range c {
			log.Println(sig)
			log.Println("Thinkpad Backlight service exiting...")
			configFile(&cfg, *configStr, "write")
			os.Exit(0)
		}
	}()
	log.Println("Thinkpad Backlight service operational")
	for {
		currentState := getPowerStatus(cfg.PowerSupplyDevPath)
		currentBrightness := getBackliteStatus(cfg.BrightnessDevPath)
		if currentState != cfg.State.PrevState {
			if currentState == "1" {
				if (cfg.State.OnAC != "") && (cfg.State.OnAC != currentBrightness) {
					setBrightness(&cfg, cfg.State.OnAC, 123)
				}
			} else {
				if (cfg.State.OnBat != "") && (cfg.State.OnBat != currentBrightness) {
					setBrightness(&cfg, cfg.State.OnBat, 124)
				}
			}
			cfg.State.PrevState = currentState
		} else {
			if currentState == "1" {
				if cfg.State.OnAC != currentBrightness {
					cfg.State.OnAC = currentBrightness
					log.Println("Updated AC Backlite to " + currentBrightness)
				}
			} else {
				if cfg.State.OnBat != currentBrightness {
					cfg.State.OnBat = currentBrightness
					log.Println("Updated BAT Backlite to " + currentBrightness)
				}
			}

		}

		time.Sleep(1 * time.Second)
	}
	select {}
}
