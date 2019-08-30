package main

import (
	"github.com/Yesterday17/go-drcom-jlu/drcom"
	"github.com/vishvananda/netlink"
	"log"
	"time"
)

// TODO: https://github.com/Microsoft/hcsshim
func initNetList(ch chan netlink.AddrUpdate) {
	list, err := netlink.LinkList()
	if err != nil {
		log.Fatal(err)
	}

	for _, link := range list {
		if link.Attrs().EncapType == "loopback" {
			continue
		}

		if link.Attrs().OperState.String() == "up" {
			mac := link.Attrs().HardwareAddr.String()
			if mac != wifiMAC {
				wiredMAC = mac
			}
		}

		done := make(chan struct{})
		if err := netlink.AddrSubscribeWithOptions(ch, done, netlink.AddrSubscribeOptions{ListExisting: false}); err != nil {
			log.Fatal(err)
		}
	}

	// 有线网络优先
	if wiredMAC != "" {
		//isWiFiMode = false
		activeMAC = wiredMAC
	}
}

func watchNetStatus(ch chan netlink.AddrUpdate) {
	plugIn := 0
	for {
		select {
		case update := <-ch:
			//mac := update
			if update.NewAddr && !isOn {
				if update.LinkAddress.IP.IsLoopback() || update.LinkAddress.IP.IsMulticast() || update.LinkAddress.IP.To4() == nil {
					continue
				}

				if plugIn < 5 {
					plugIn++
					continue
				}

				// TODO: 检查网络可用性

				log.Printf("[GDJ][Info] Detected network connected, connecting...")
				time.Sleep(time.Second * 2) // sleep 2 seconds to connect
				client = drcom.New(cfg)
				client.Start()
				isOn = true
			} else if !update.NewAddr && isOn {
				_ = client.Close()
				log.Printf("[GDJ][Info] Network disconnected.")
				isOn = false
			}
		}
	}
}
