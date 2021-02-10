package cherryTime

import "time"

var (
	offsetSecond int64

	YYYYMMDDHHMMSS = "2006-01-02 15:04:05"
)

func AddOffset(second int64) {
	offsetSecond = second
}

func SubOffset(second int64) {
	offsetSecond = -second
}

func NowSecond() int64 {
	return time.Now().Unix() + offsetSecond
}

func NowMillisecond() int64 {
	return time.Now().UnixNano()/1e6 + offsetSecond*1000
}

func UnixTimeToString(second int64) string {
	return UnixFormat(second, YYYYMMDDHHMMSS)
}

func UnixFormat(second int64, format string) string {
	return time.Unix(second, 0).Format(format)
}
