package cherryNats

import (
	"strconv"
	"sync/atomic"
)

var (
	atomicReqID = &atomic.Uint64{}
)

func NewReqID() uint64 {
	return atomicReqID.Add(1)
}

func NewStringReqID() string {
	return strconv.FormatUint(NewReqID(), 10)
}
