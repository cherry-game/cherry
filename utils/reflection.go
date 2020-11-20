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
