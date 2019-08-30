package main

import (
	"fmt"
	"github.com/Yesterday17/go-drcom-jlu/config"
	"github.com/Yesterday17/go-drcom-jlu/drcom"
	"github.com/vishvananda/netlink"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	isOn      = false
	activeMAC = ""
	//isWiFiMode = false
	wifiMAC  = ""
	wiredMAC = ""
	client   *drcom.Service
	cfg      *config.Config
)

// return code list
// -10 failed to parse config file

func main() {
	var err error

	if ssid, inf, err := getWifiName(); err != nil {
		fmt.Println(err)
	} else if ssid == "JLU.PC" {
		//isWiFiMode = true
		wifiMAC = inf.HardwareAddr.String()
		activeMAC = wifiMAC
		fmt.Println("[GDJ][INFO] Wireless network JLU.PC detected")
		fmt.Println("[GDJ][INFO] Wireless MAC address: " + wifiMAC)
	} else if ssid == "" {
		// not connected
		//fmt.Println(inf)
	} else {
		wifiMAC = inf.HardwareAddr.String()
	}

	ch := make(chan netlink.AddrUpdate)
	initNetList(ch)

	// 加载配置文件
	conf, err := config.ReadConfig("./config.json")
	// conf, err := config.ReadConfig("/etc/go-drcom-jlu/config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(-10)
	}

	// 检查配置文件的 MAC 地址是否与 WiFi / 有线网卡 的 MAC 匹配
	//if conf.MAC != wiredMAC && conf.MAC != activeMAC {
	//	fmt.Println("[GDJ][ERROR] Invalid MAC address")
	//	os.Exit(-10)
	//}

	// 全局化
	cfg = &conf

	go watchNetStatus(ch)

	if activeMAC != "" {
		client = drcom.New(cfg)
		client.Start()
		isOn = true
	}

	// 处理退出信号
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-sigc
		log.Printf("[GDJ] Exiting with signal %s", s.String())
		if client != nil {
			client.Logout()
			_ = client.Close()
		}
		return
	}
}
