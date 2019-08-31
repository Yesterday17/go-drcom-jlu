package main

import (
	"flag"
	"fmt"
	"github.com/Yesterday17/go-drcom-jlu/drcom"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	activeMAC = ""
	client    *drcom.Service
	cfg       *drcom.Config
)

// return code list
// -10 failed to parse config file

func main() {
	var cfgPath string
	var err error

	flag.StringVar(&cfgPath, "c", "./config.json", "配置文件的路径")
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
		fmt.Println(err)
		os.Exit(-10)
	}

	// 检查配置文件的 MAC 地址是否与 WiFi / 有线网卡 的 MAC 匹配
	for _, inf := range Interfaces {
		if inf.Address == cfg.MAC {
			if inf.IsWireless {
				fmt.Printf("[GDJ][WARN] Wireless MAC address detected")
			}
			activeMAC = inf.Address
			break
		}
	}

	// 未检测到对应配置文件的 MAC 地址
	if activeMAC == "" {
		log.Fatal("[GDJ][ERROR] No matching MAC address detected")
	} else {
		inf := Interfaces[activeMAC]
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
		log.Printf("[GDJ] Exiting with signal %s", s.String())
		if client != nil {
			client.Logout()
			_ = client.Close()
		}
		return
	}
}
