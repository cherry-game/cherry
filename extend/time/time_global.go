package cherryTime

import "time"

var (
	globalOffsetSecond int            //全局偏移值(秒)
	globalLocation     *time.Location //全局时区
)

func init() {
	SetGlobalLocation("Local")
}

func AddGlobalOffset(second int) {
	globalOffsetSecond = second
}

func SubGlobalOffset(second int) {
	globalOffsetSecond = -second
}

func SetGlobalLocation(name string) (err error) {
	globalLocation, err = time.LoadLocation(name)
	if err != nil {
		return err
	}

	return nil
}
