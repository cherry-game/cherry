package cherryHandler

import (
	cfacade "github.com/cherry-game/cherry/facade"
	"math/rand"
	"sync/atomic"
)

var (
	roundAtomicId int64 = 0
)

// RandomQueueHash 随机handler queue
func RandomQueueHash(_ cfacade.IExecutor, queueNum int) int {
	return rand.Intn(queueNum)
}

// RoundQueueHash 轮询handler queue
func RoundQueueHash(_ cfacade.IExecutor, queueNum int) int {
	atomicId := atomic.AddInt64(&roundAtomicId, 1)
	return int(atomicId % int64(queueNum))
}
