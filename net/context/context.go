package cherryContext

import (
	"context"
	"encoding/json"
	cconst "github.com/cherry-game/cherry/const"
)

const (
	zeroUint uint = 0
)

// Add adds a key and value that will be propagated through RPC calls
func Add(ctx context.Context, key string, val interface{}) context.Context {
	propagate := ToMap(ctx)
	propagate[key] = val
	return context.WithValue(ctx, cconst.PropagateCtxKey, propagate)
}

// Get get a value from the propagate
func Get(ctx context.Context, key string) interface{} {
	propagate := ToMap(ctx)
	if val, ok := propagate[key]; ok {
		return val
	}
	return nil
}

func GetMessageId(ctx context.Context) uint {
	val := Get(ctx, cconst.MessageIdKey)
	if val == nil {
		return zeroUint
	}

	if i, ok := val.(uint); ok {
		return i
	}
	return zeroUint
}

// ToMap returns the values that will be propagated through RPC calls in map[string]interface{} format
func ToMap(ctx context.Context) map[string]interface{} {
	if ctx == nil {
		return map[string]interface{}{}
	}
	p := ctx.Value(cconst.PropagateCtxKey)
	if p != nil {
		return p.(map[string]interface{})
	}
	return map[string]interface{}{}
}

// FromMap creates a new context from a map with propagated values
func FromMap(val map[string]interface{}) context.Context {
	return context.WithValue(context.Background(), cconst.PropagateCtxKey, val)
}

// Encode returns the given propagatable context encoded in binary format
func Encode(ctx context.Context) ([]byte, error) {
	m := ToMap(ctx)
	if len(m) > 0 {
		return json.Marshal(m)
	}
	return nil, nil
}

// Decode returns a context given a binary encoded message
func Decode(m []byte) (context.Context, error) {
	if len(m) == 0 {
		// TODO maybe return an error
		return nil, nil
	}
	mp := make(map[string]interface{}, 0)
	err := json.Unmarshal(m, &mp)
	if err != nil {
		return nil, err
	}
	return FromMap(mp), nil
}
