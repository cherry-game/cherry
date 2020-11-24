package cherryUtils

import (
	goNet "net"
	"sync"
)

var localIPv4Str = "0.0.0.0"
var localIPv4Once = new(sync.Once)

type net struct {
}

func (n *net) LocalIPV4() string {
	localIPv4Once.Do(func() {
		if ias, err := goNet.InterfaceAddrs(); err == nil {
			for _, address := range ias {
				if ipNet, ok := address.(*goNet.IPNet); ok && !ipNet.IP.IsLoopback() {
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
