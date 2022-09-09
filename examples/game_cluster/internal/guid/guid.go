package guid

import "go.uber.org/atomic"

var (
	nextId = &atomic.Int64{}
)

// Next 生成唯一id
// TODO 本guid生成仅做演示用，正式环境可以使用其他方式生成全局唯一id
// 以下几种方式仅供参考：
// snowflake
// redis
func Next() int64 {
	return nextId.Inc()
}
