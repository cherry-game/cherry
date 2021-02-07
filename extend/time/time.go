package cherryTime

import "time"

func NowMillisecond() int {
	return int(time.Now().UnixNano() / 1000000)
}

func UnixTimeToString(unixTime int64) string {
	return time.Unix(unixTime, 0).Format("2006-01-02 15:04:05")
}
