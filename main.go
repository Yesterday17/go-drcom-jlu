package main

import (
	"fmt"
	"github.com/Yesterday17/go-drcom-jlu/config"
	"github.com/Yesterday17/go-drcom-jlu/drcom"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	activeMAC = ""
	client    *drcom.Service
	cfg       *config.Config
)

// return code list
// -10 failed to parse config file

func main() {
	Interfaces = make(map[string]*Interface)

	if err := initWireless(); err != nil {
		log.Fatal(err)
	}

	if err := initWired(); err != nil {
		log.Fatal(err)
	}

	// 加载配置文件
	conf, err := config.ReadConfig("./config.json")
	// conf, err := config.ReadConfig("/etc/go-drcom-jlu/config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(-10)
	}

	// 检查配置文件的 MAC 地址是否与 WiFi / 有线网卡 的 MAC 匹配
	for _, inf := range Interfaces {
		if inf.Address == conf.MAC {
			if inf.IsWireless {
				fmt.Printf("[GDJ][WARN] Wireless MAC address detected, make sure you know what you're doing!")
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

	// 全局化
	cfg = &conf

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
