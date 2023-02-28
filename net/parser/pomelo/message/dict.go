package pomeloMessage

import (
	clog "github.com/cherry-game/cherry/logger"
	"strings"
)

var (
	routes = make(map[string]uint16) // 路由信息映射为uint16
	codes  = make(map[uint16]string) // uint16映射为路由信息
)

// SetDictionary set routes map which be used to compress route.
func SetDictionary(dict map[string]uint16) {
	if dict == nil {
		return
	}

	for route, code := range dict {
		r := strings.TrimSpace(route) //去掉开头结尾的空格

		// duplication check
		if _, ok := routes[r]; ok {
			clog.Errorf("duplicated route(route: %s, code: %d)", r, code)
			return
		}

		if _, ok := codes[code]; ok {
			clog.Errorf("duplicated route(route: %s, code: %d)", r, code)
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

func GetCode(route string) (uint16, bool) {
	code, found := routes[route]
	return code, found
}
