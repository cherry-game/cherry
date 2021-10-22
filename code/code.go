package cherryCode

import (
	cherryProto "github.com/cherry-game/cherry/net/proto"
	"sync"
)

const (
	emptyMsg = ""
)

var (
	lock           = &sync.RWMutex{}
	codeResultMaps = make(map[int32]*cherryProto.CodeResult)
)

func Add(code int32, message string) {
	lock.Lock()
	defer lock.Unlock()

	codeResultMaps[code] = &cherryProto.CodeResult{
		Code:    code,
		Message: message,
	}
}

func AddAll(maps map[int32]string) {
	for k, v := range maps {
		Add(k, v)
	}
}

func GetCodeResult(code int32) *cherryProto.CodeResult {
	val, found := codeResultMaps[code]
	if found {
		return val
	}

	Add(code, emptyMsg)
	return GetCodeResult(code)
}

func GetMessage(code int32) string {
	statusCode, found := codeResultMaps[code]
	if found {
		return statusCode.Message
	}

	return emptyMsg
}

func IsOK(code int32) bool {
	return code == OK
}

func IsFail(code int32) bool {
	return code > OK
}
