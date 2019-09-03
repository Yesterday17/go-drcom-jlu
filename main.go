package main

import (
	"flag"
	"github.com/Yesterday17/go-drcom-jlu/drcom"
	"github.com/Yesterday17/go-drcom-jlu/logger"
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
	var cfgPath string
	var err error

	flag.StringVar(&cfgPath, "config", "./config.json", "配置文件的路径")
	flag.Parse()

	Interfaces = make(map[string]*Interface)

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
	var MAC string
	for _, inf := range Interfaces {
		if inf.Address == cfg.MAC {
			if inf.IsWireless {
				logger.Warn("Wireless MAC address detected")
			}
			MAC = inf.Address
			break
		}
	}

	// 未检测到对应配置文件的 MAC 地址
	if MAC == "" {
		logger.Error("No matching MAC address detected")
		os.Exit(10)
	} else {
		inf := Interfaces[MAC]
		if !inf.IsSchoolNet() {
			MAC = ""
		}

		// 当 MAC 对应的接口未连接时 搜索无线网卡
		if !inf.Connected {
			for _, inf2 := range Interfaces {
				if inf2.IsWireless && inf2.IsSchoolNet() {
					MAC = inf2.Address
					break
				}
			}
		}
	}

	go watchNetStatus()

	if MAC != "" {
		NewClient(MAC)
	}

	// 处理退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-sig
		log.Printf("Exiting with signal %s", s.String())
		if activeMAC != "" && client != nil {
			// client.Logout()
			_ = client.Close()
		}
		return
	}
}
