package code

import (
	cherryCode "github.com/cherry-game/cherry/code"
	"sync"
)

const (
	emptyMsg = ""
)

var (
	lock       = &sync.RWMutex{}
	resultMaps = make(map[int32]string)
)

func AddAll(maps map[int32]string) {
	for k, v := range maps {
		Add(k, v)
	}
}

func Add(code int32, message string) {
	lock.Lock()
	defer lock.Unlock()
	resultMaps[code] = message
}

func GetMessage(code int32) string {
	msg, found := resultMaps[code]
	if found {
		return msg
	}

	return emptyMsg
}

func IsOK(code int32) bool {
	return cherryCode.IsOK(code)
}

func IsFail(code int32) bool {
	return cherryCode.IsFail(code)
}
