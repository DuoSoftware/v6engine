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

	addrl, _ := net.LookupAddr("0.0.0.0")

	return addrl[0]
}
