package main

import (
	"fmt"
	"github.com/Yesterday17/go-drcom-jlu/drcom"
	"github.com/Yesterday17/go-drcom-jlu/logger"
	"github.com/mdlayher/wifi"
	"github.com/vishvananda/netlink"
	"log"
	"time"
)

type Interface struct {
	Name      string
	Address   string
	Connected bool

	IsSchoolNet bool

	IsWireless bool
	SSID       string
}

// MAC -> Interface 的 Map
var Interfaces map[string]*Interface

// TODO: 支持 Windows
func initWireless() error {
	if client, err := wifi.New(); err != nil {
		return err
	} else {
		defer client.Close()
		interfaces, err := client.Interfaces()
		if err != nil {
			return err
		}

		for _, inf := range interfaces {
			if inf.Name == "" {
				continue
			}

			MAC := inf.HardwareAddr.String()
			_interface := &Interface{
				Name:       inf.Name,
				Address:    MAC,
				Connected:  false,
				IsWireless: true,
				SSID:       "",
			}
			Interfaces[MAC] = _interface

			bss, err := client.BSS(inf)
			if err != nil {
				continue
			}
			_interface.Connected = true
			_interface.SSID = bss.SSID
		}
	}
	return nil
}

// TODO: https://github.com/Microsoft/hcsshim
func initWired() error {
	list, err := netlink.LinkList()
	if err != nil {
		return err
	}

	for _, link := range list {
		// 避免初始化环回接口
		if link.Attrs().EncapType == "loopback" {
			continue
		}

		MAC := link.Attrs().HardwareAddr.String()
		if Interfaces[MAC] != nil {
			// WiFi 网络已初始化 跳过
			continue
		}

		// 设置有线网络
		inf := &Interface{
			Name:       link.Attrs().Name,
			Address:    MAC,
			Connected:  false,
			IsWireless: false,
			SSID:       "",
		}
		Interfaces[MAC] = inf
		inf.Connected = link.Attrs().OperState.String() == "up"
	}
	return nil
}

func getSSID(MAC string) (string, error) {
	if client, err := wifi.New(); err != nil {
		return "", err
	} else {
		defer client.Close()
		interfaces, err := client.Interfaces()
		if err != nil {
			return "", err
		}

		for _, inf := range interfaces {
			if inf.HardwareAddr.String() != MAC || inf.Name == "" {
				continue
			}

			bss, err := client.BSS(inf)
			if err != nil {
				return "", nil
			}

			return bss.SSID, nil
		}
		return "", fmt.Errorf("failed to get ssid")
	}
}

func watchNetStatus() {
	// 建立 ch
	ch := make(chan netlink.LinkUpdate)

	// 注册监听 Addr 变化
	done := make(chan struct{})
	if err := netlink.LinkSubscribe(ch, done); err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case update := <-ch:
			if update.Attrs().MTU > 0 &&
				update.Attrs().OperState.String() == "up" &&
				update.Flags >= 65536 &&
				activeMAC == "" {

				MAC := update.Attrs().HardwareAddr.String()

				inf := Interfaces[MAC]
				if inf == nil {
					continue
				}

				if inf.IsWireless {
					if ssid, err := getSSID(inf.Address); err != nil {
						logger.Error("Failed to get SSID of connecting WiFi")
						continue
					} else {
						inf.SSID = ssid
						if ssid != "JLU.PC" {
							logger.Info("Skipping non-JLU.PC WiFI")
							continue
						}
					}
				}

				// TODO: 检查网络可用性
				logger.Info("Network connected, connecting...")
				time.Sleep(time.Second * 2)
				client = drcom.New(cfg)
				client.Start()

				activeMAC = MAC
				inf.Connected = true
			} else if update.Flags < 65536 &&
				activeMAC != "" &&
				update.Attrs().Name != "" &&
				update.Attrs().HardwareAddr.String() == activeMAC {
				MAC := update.Attrs().HardwareAddr.String()
				inf := Interfaces[MAC]
				if inf == nil {
					continue
				}

				_ = client.Close()
				logger.Info("Network disconnected")

				activeMAC = ""
				inf.Connected = false
				inf.SSID = ""
			}
		}
	}
}
