package cherryContext

//import (
//	"context"
//	cstring "github.com/cherry-game/cherry/extend/string"
//)
//
//func New() context.Context {
//	return context.TODO()
//}
//
//// Add adds a key and value that will be propagated through RPC calls
//func Add(ctx context.Context, key string, val interface{}) context.Context {
//	propagate := ToMap(ctx)
//	propagate[key] = cstring.ToString(val)
//	return context.WithValue(ctx, PropagateCtxKey, propagate)
//}
//
//func ParseToString(ctx context.Context, key string) string {
//	propagate := ToMap(ctx)
//	if val, ok := propagate[key]; ok {
//		return val
//	}
//	return ""
//}
//
//func PraseToInt(ctx context.Context, key string) int {
//	v := ParseToString(ctx, key)
//	if v == "" {
//		return 0
//	}
//
//	value, ok := cstring.ToInt(v)
//	if !ok {
//		return 0
//	}
//	return value
//}
//
//// ParseToInt32 returns the value associated with the key as a int32.
//func ParseToInt32(ctx context.Context, key string) int32 {
//	v := ParseToString(ctx, key)
//	if v == "" {
//		return 0
//	}
//
//	value, ok := cstring.ToInt32(v)
//	if !ok {
//		return 0
//	}
//	return value
//}
//
//// ParseToInt64 returns the value associated with the key as a int64.
//func ParseToInt64(ctx context.Context, key string) int64 {
//	v := ParseToString(ctx, key)
//	if v == "" {
//		return 0
//	}
//
//	value, ok := cstring.ToInt64(v)
//	if !ok {
//		return 0
//	}
//	return value
//}
//
//func ParseToUint(ctx context.Context, key string) uint {
//	v := ParseToString(ctx, key)
//	if v == "" {
//		return 0
//	}
//
//	value, ok := cstring.ToUint(v, 0)
//	if ok {
//		return value
//	}
//	return 0
//}
//
//// ToMap returns the values that will be propagated through RPC calls in map[string]interface{} format
//func ToMap(ctx context.Context) map[string]string {
//	if ctx == nil {
//		return map[string]string{}
//	}
//
//	p := ctx.Value(PropagateCtxKey)
//	if p != nil {
//		return p.(map[string]string)
//	}
//
//	return map[string]string{}
//}
//
//func FromMap(val map[string]string) context.Context {
//	return context.WithValue(context.Background(), PropagateCtxKey, val)
//}
