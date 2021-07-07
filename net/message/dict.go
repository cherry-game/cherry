package cherryMessage

import (
	cherryLogger "github.com/cherry-game/cherry/logger"
	"strings"
)

var (
	routes = make(map[string]uint16) // 路由信息映射为uint16
	codes  = make(map[uint16]string) // uint16映射为路由信息
)

// 启用路由压缩
// 对于服务端，server会扫描所有的Handler信息
// 对于客户端，用户需要配置一个路由映射表
// 通过这两种方式，pitaya会拿到所有的客户端和服务端的路由信息，然后将每一个路由信息都映射为一个小整数，
// 在客户端与服务器建立连接的握手过程中，服务器会将 整个字典传给客户端，
// 这样在以后的通信中，对于路由信息，将全部使用定义的小整数进行标记，大大地减少了额外信 息开销
// SetDictionary set routes map which be used to compress route.
func SetDictionary(dict map[string]uint16) {
	if dict == nil {
		return
	}

	for route, code := range dict {
		r := strings.TrimSpace(route) //去掉开头结尾的空格

		// duplication check
		if _, ok := routes[r]; ok {
			cherryLogger.Errorf("duplicated route(route: %s, code: %d)", r, code)
			return
		}

		if _, ok := codes[code]; ok {
			cherryLogger.Errorf("duplicated route(route: %s, code: %d)", r, code)
			return
		}

		// update map, using last value when key duplicated
		routes[r] = code
		codes[code] = r
	}

	return
}

// GetDictionary gets the routes map which is used to compress route.
func GetDictionary() map[string]uint16 {
	return routes
}

func GetRoute(code uint16) (route string, found bool) {
	route, found = codes[code]
	return route, found
}
