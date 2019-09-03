package main

import (
	"encoding/json"
	"fmt"
	"github.com/Yesterday17/go-drcom-jlu/drcom"
	"github.com/Yesterday17/go-drcom-jlu/logger"
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

	// Default LogLevel = 1
	if config.LogLevel < 0 || config.LogLevel > 2 {
		config.LogLevel = 1
	}

	if config.LogPath == "" {
		switch config.LogLevel {
		case 0:
			logger.Init(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
		case 1:
			logger.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
		case 2:
			logger.Init(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
		}
	} else {
		// TODO: 写入日志到文件
		switch config.LogLevel {
		case 0:
			logger.Init(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
		case 1:
			logger.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
		case 2:
			logger.Init(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
		}
	}

	// Write change to config file
	jsonConfig, _ := json.Marshal(config)
	if err := ioutil.WriteFile(path, jsonConfig, os.FileMode(0644)); err != nil {
		return nil, err
	}

	return &config, nil
}
