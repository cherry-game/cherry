package cherryNet

import (
	"net"
	"sync"
)

var localIPv4Str = "0.0.0.0"
var localIPv4Once = new(sync.Once)

func LocalIPV4() string {
	localIPv4Once.Do(func() {
		if ias, err := net.InterfaceAddrs(); err == nil {
			for _, address := range ias {
				if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
					if ipNet.IP.To4() != nil {
						localIPv4Str = ipNet.IP.String()
						return
					}
				}
			}
		}
	})
	return localIPv4Str
}

func GetIPV4(addr net.Addr) string {
	if addr == nil {
		return ""
	}

	if ipNet, ok := addr.(*net.TCPAddr); ok {
		return ipNet.IP.String()
	}

	if ipNet, ok := addr.(*net.UDPAddr); ok {
		return ipNet.IP.String()
	}

	return ""
}
