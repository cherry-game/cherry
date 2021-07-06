package main

import (
	"fmt"
	"reflect"
)

type helloP1 struct {
	a int
	b string
}

func (h *helloP1) ToString() string {
	return fmt.Sprintf("a=%d,b=%s", h.a, h.b)
}

func hello(x helloP1) {
	fmt.Println("Hello world!")
	fmt.Println(x.ToString())
}

type InvokeFunc struct {
	Type    reflect.Type
	Value   reflect.Value
	InArgs  []reflect.Type
	OutArgs []reflect.Type
}

func main() {
	hl := hello
	execfn(hl)
}

func execfn(hl interface{}) {
	fv := reflect.ValueOf(hl)
	fmt.Println("fv is reflect.Func ?", fv.Kind() == reflect.Func)

	invokeFunc := &InvokeFunc{Value: fv}

	params := make([]reflect.Value, 1)

	p := helloP1{
		a: 1,
		b: "11111",
	}

	params[0] = reflect.ValueOf(p)
	fv.Call(params)

	invokeFunc.Value.Call(params)
}
