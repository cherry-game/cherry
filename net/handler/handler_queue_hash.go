package cherryHandler

import (
	cfacade "github.com/cherry-game/cherry/facade"
	"math/rand"
	"sync/atomic"
)

var (
	roundAtomicId = atomic.Int64{}
)

// RandomQueueHash 随机handler queue
func RandomQueueHash(_ cfacade.IExecutor, queueNum int) int {
	return rand.Intn(queueNum)
}

// RoundQueueHash 轮询handler queue
func RoundQueueHash(_ cfacade.IExecutor, queueNum int) int {
	return int(roundAtomicId.Add(1) % int64(queueNum))
}
