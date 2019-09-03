package main

import (
	"fmt"
	"github.com/Yesterday17/go-drcom-jlu/drcom"
	"github.com/Yesterday17/go-drcom-jlu/logger"
	"github.com/mdlayher/wifi"
	"github.com/vishvananda/netlink"
	"log"
)

type Interface struct {
	// 初始化后不应改变的值
	Name       string
	Address    string
	IsWireless bool

	// 随接口状态变化的值
	Connected   bool
	SSID        string
	CanPingUIMS bool
}

func (i *Interface) IsSchoolNet() bool {
	if !i.Connected {
		return false
	}

	if i.IsWireless {
		return i.SSID == "JLU.PC"
	} else {
		return true
		// return i.CanPingUIMS
	}
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
				update.Attrs().Name != "" &&
				update.Flags >= 65536 {
				// 更新接口状态
				MAC := update.Attrs().HardwareAddr.String()
				inf := Interfaces[MAC]
				if inf == nil {
					continue
				}

				// 标记接口为已连接
				inf.Connected = true

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
				if activeMAC == "" {
					logger.Infof("%v", update)
					logger.Info("Network connected, connecting...")
					NewClient(MAC)
				}
			} else if update.Flags < 65536 &&
				update.Attrs().Name != "" {
				// 更新接口状态
				MAC := update.Attrs().HardwareAddr.String()
				inf := Interfaces[MAC]
				if inf == nil {
					continue
				}
				inf.Connected = false

				if activeMAC != "" &&
					update.Attrs().HardwareAddr.String() == activeMAC {
					_ = client.Close()
					activeMAC = ""
					inf.SSID = ""
					logger.Info("Network disconnected")

					// 寻找其他已接入网络
					for _, inf := range Interfaces {
						if inf.IsSchoolNet() {
							NewClient(inf.Address)
							break
						}
					}
				}
			}
		}
	}
}

func NewClient(MAC string) {
	inf := Interfaces[MAC]
	logger.Infof("Connecting with interface %s - %s", inf.Name, inf.Address)

	activeMAC = MAC
	client = drcom.New(cfg)
	client.Start()
}
