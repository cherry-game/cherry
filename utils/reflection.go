package cherryUtils

import (
	"fmt"
	"reflect"
	"runtime"
)

type reflection struct {
}

func (r *reflection) ReflectTry(f reflect.Value, args []reflect.Value, handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("-------------panic recover---------------")
			if handler != nil {
				handler(err)
			}
		}
	}()
	f.Call(args)
}

func (r *reflection) GetStructName(v interface{}) string {
	return reflect.Indirect(reflect.ValueOf(v)).Type().Name()
}

func (r *reflection) GetFuncName(fn interface{}) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	return Strings.CutLastString(fullName, ".", "-")
}

// IsNil 返回 reflect.Value 的值是否为 nil，比原生方法更安全
func (r *reflection) IsNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return v.IsNil()
	}
	return false
}
