package cherryContext

import (
	"context"
	clog "github.com/cherry-game/cherry/logger"
	jsoniter "github.com/json-iterator/go"
)

const (
	zeroUint  uint  = 0
	zeroInt64 int64 = 0
)

func New() context.Context {
	ctx := context.TODO()
	return ctx
}

// Add adds a key and value that will be propagated through RPC calls
func Add(ctx context.Context, key string, val interface{}) context.Context {
	propagate := ToMap(ctx)
	propagate[key] = val
	return context.WithValue(ctx, PropagateCtxKey, propagate)
}

// Get get a value from the propagate
func Get(ctx context.Context, key string) interface{} {
	propagate := ToMap(ctx)
	if val, ok := propagate[key]; ok {
		return val
	}
	return nil
}

func GetInt64(ctx context.Context, key string) int64 {
	val := Get(ctx, key)
	if val == nil {
		return zeroInt64
	}

	f64, ok := val.(float64)
	if ok {
		return int64(f64)
	}

	i64, ok := val.(int64)
	if ok {
		return i64
	}

	return zeroInt64
}

func GetMessageId(ctx context.Context) uint {
	val := Get(ctx, MessageIdKey)
	if val == nil {
		return zeroUint
	}

	i, ok := val.(float64)
	if ok {
		return uint(i)
	}

	return zeroUint
}

func GetBuildPacketTime(ctx context.Context) int64 {
	return GetInt64(ctx, BuildPacketTimeKey)
}

func GetInHandlerTime(ctx context.Context) int64 {
	return GetInt64(ctx, InHandlerTimeKey)
}

// ToMap returns the values that will be propagated through RPC calls in map[string]interface{} format
func ToMap(ctx context.Context) map[string]interface{} {
	if ctx == nil {
		return map[string]interface{}{}
	}
	p := ctx.Value(PropagateCtxKey)
	if p != nil {
		return p.(map[string]interface{})
	}
	return map[string]interface{}{}
}

// FromMap creates a new context from a map with propagated values
func FromMap(val map[string]interface{}) context.Context {
	return context.WithValue(context.Background(), PropagateCtxKey, val)
}

// Encode returns the given propagatable context encoded in binary format
func Encode(ctx context.Context) []byte {
	m := ToMap(ctx)
	if len(m) > 0 {
		bytes, err := jsoniter.Marshal(m)
		if err != nil {
			clog.Warn(err)
		}
		return bytes
	}
	return nil
}

// Decode returns a context given a binary encoded message
func Decode(m []byte) context.Context {
	if len(m) == 0 {
		return context.TODO()
	}
	mp := make(map[string]interface{}, 0)
	err := jsoniter.Unmarshal(m, &mp)
	if err != nil {
		return context.TODO()
	}

	return FromMap(mp)
}
