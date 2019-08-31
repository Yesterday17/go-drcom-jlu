package main

import (
	"flag"
	"github.com/Yesterday17/go-drcom-jlu/drcom"
	"github.com/Yesterday17/go-drcom-jlu/logger"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	activeMAC = ""
	client    *drcom.Client
	cfg       *drcom.Config
)

// return code list
// 10 failed to parse config file

func main() {
	var cfgPath, logPath string
	var logLevel int
	var err error

	flag.StringVar(&cfgPath, "c", "./config.json", "配置文件的路径")
	flag.StringVar(&logPath, "log", "", "日志文件的路径, 留空则输出到 stdout")
	flag.IntVar(&logLevel, "level", 1, "日志级别, 0 最简略, 2 最详细")
	flag.Parse()

	Interfaces = make(map[string]*Interface)

	if logLevel > 2 || logLevel < 0 {
		log.Fatalln("日志等级必须在 0-2 之间")
	}

	if logPath == "" {
		switch logLevel {
		case 0:
			logger.Init(ioutil.Discard, ioutil.Discard, os.Stderr)
		case 1:
			logger.Init(ioutil.Discard, os.Stdout, os.Stderr)
		case 2:
			logger.Init(os.Stdout, os.Stdout, os.Stderr)
		}
	} else {
		// TODO: 写入日志到文件
	}

	if err = initWireless(); err != nil {
		log.Fatal(err)
	}

	if err = initWired(); err != nil {
		log.Fatal(err)
	}

	// 加载配置文件
	cfg, err = ReadConfig(cfgPath)
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	// 检查配置文件的 MAC 地址是否与 WiFi / 有线网卡 的 MAC 匹配
	for _, inf := range Interfaces {
		if inf.Address == cfg.MAC {
			if inf.IsWireless {
				logger.Warn("Wireless MAC address detected")
			}
			activeMAC = inf.Address
			break
		}
	}

	// 未检测到对应配置文件的 MAC 地址
	if activeMAC == "" {
		logger.Error("No matching MAC address detected")
		os.Exit(1)
	} else {
		inf := Interfaces[activeMAC]

		if !inf.Connected {
			for _, inf2 := range Interfaces {
				if inf2.IsWireless && inf2.Connected {
					inf = inf2
					activeMAC = inf2.Address
				}
			}
		}

		if !inf.Connected {
			activeMAC = ""
		} else if inf.IsWireless && inf.SSID != "JLU.PC" {
			activeMAC = ""
		}
	}

	go watchNetStatus()

	if activeMAC != "" {
		client = drcom.New(cfg)
		client.Start()
	}

	// 处理退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-sig
		log.Printf("Exiting with signal %s", s.String())
		if client != nil {
			client.Logout()
			_ = client.Close()
		}
		return
	}
}
