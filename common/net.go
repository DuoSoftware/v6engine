package common

import (
	//"fmt"
	"net"
)

func GetLocalHostName() (name string) {
	/*
		ifaces, _ := net.Interfaces()

		for _, i := range ifaces {
			addrs, _ := i.Addrs()
			// handle err
			for _, addr := range addrs {
				switch v := addr.(type) {
				case *net.IPAddr:
					fmt.Println(addr.String())
					_ = v
					addrl, _ := net.LookupAddr("0.0.0.0")
					fmt.Println(addrl)
					//return v.String() //addr.String()
				}

			}
		}*/

	/*
		addrl, _ := net.LookupAddr("0.0.0.0")
		if len(addrl) == 0 {
			return "0.0.0.0"
		} else {
			return addrl[0]
		}
	*/

	ifaces, err := net.Interfaces()
	if err != nil {
		return "UnknownHost(I)"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "UnknownHost(A)"
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String()
		}
	}

	return "UnknownHost"
}
