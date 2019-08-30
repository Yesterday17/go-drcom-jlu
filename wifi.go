package main

import "github.com/mdlayher/wifi"

func getWifiName() (string, *wifi.Interface, error) {
	var inf *wifi.Interface
	ssid := ""
	if client, err := wifi.New(); err != nil {
		return "", nil, err
	} else {
		defer client.Close()
		interfaces, err := client.Interfaces()
		if err != nil {
			return "", nil, err
		}

		for _, _interface := range interfaces {
			bss, err := client.BSS(_interface)
			if err != nil {
				inf = nil
				continue
			}
			inf = _interface
			ssid = bss.SSID
		}
	}
	return ssid, inf, nil
}
