package main

import (
	"encoding/json"
	"fmt"
	"github.com/Yesterday17/go-drcom-jlu/drcom"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func ReadConfig(path string) (*drcom.Config, error) {
	var config drcom.Config

	if content, err := ioutil.ReadFile(path); err != nil {
		return nil, err
	} else if err = json.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	if match, _ := regexp.MatchString("(?:[0-9A-Za-z]{2}:){5}[0-9A-Za-z]{2}", config.MAC); !match {
		return nil, fmt.Errorf("invalid MAC address")
	}

	if match, _ := regexp.MatchString("^[a-z]{4,}\\d{4}$", config.Username); !match {
		return nil, fmt.Errorf("invalid username")
	}

	if config.Password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	// convert MAC to lower case
	config.MAC = strings.ToLower(config.MAC)

	// Default Timeout = 3 seconds
	if config.Timeout <= 0 {
		config.Timeout = 3
	}

	// Default Retry = 3 times
	if config.Retry <= 0 {
		config.Retry = 3
	}

	// Write change to config file
	jsonConfig, _ := json.Marshal(config)
	if err := ioutil.WriteFile(path, jsonConfig, os.FileMode(0644)); err != nil {
		return nil, err
	}

	return &config, nil
}
