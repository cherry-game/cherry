package cherryUtils

import (
	"fmt"
	cherryInterfaces "github.com/cherry-game/cherry/interfaces"
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

//getInvokeFunc reflect function convert to InvokeFn
func (r *reflection) GetInvokeFunc(name string, fn interface{}) (*cherryInterfaces.InvokeFn, error) {
	if name == "" {
		return nil, Error("func name is nil")
	}

	if fn == nil {
		return nil, Errorf("func is nil. name = %s", name)
	}

	typ := reflect.TypeOf(fn)
	val := reflect.ValueOf(fn)

	if typ.Kind() != reflect.Func {
		return nil, Errorf("name = %s is not func type.", name)
	}

	var inArgs []reflect.Type
	for i := 0; i < typ.NumIn(); i++ {
		t := typ.In(i)
		inArgs = append(inArgs, t)
	}

	var outArgs []reflect.Type
	for i := 0; i < typ.NumOut(); i++ {
		t := typ.Out(i)
		outArgs = append(outArgs, t)
	}

	invoke := &cherryInterfaces.InvokeFn{
		Type:    typ,
		Value:   val,
		InArgs:  inArgs,
		OutArgs: outArgs,
	}

	return invoke, nil
}
