package cherryTime

import (
	"sync/atomic"
	"time"
)

var (
	offsetTime     atomic.Int64  //全局偏移时间
	offsetLocation *time.Location //全局偏移时区
)

func init() {
	SetOffsetLocation("Local")
}

func AddOffsetTime(t time.Duration) {
	offsetTime.Store(int64(t))
}

func SubOffsetTime(t time.Duration) {
	offsetTime.Store(int64(-t))
}

func SetOffsetLocation(name string) (err error) {
	offsetLocation, err = time.LoadLocation(name)
	if err != nil {
		return err
	}

	return nil
}
