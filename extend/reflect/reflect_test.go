package cherryReflect

import (
	"fmt"
	"reflect"
	"testing"
)

type session interface {
	Send(bytes []byte)
}

type userSession struct {
}

func (u *userSession) Send(bytes []byte) {
	fmt.Println(fmt.Sprintf("send bytes = %v", bytes))
}

type message interface {
	Data() string
}

type userMessage struct {
}

func (u *userMessage) Data() string {
	return "userMessage"
}

func abc(s session, m message) {
	fmt.Println("call session Send()")
	s.Send([]byte("bytes testing...."))

	fmt.Println("call message Msg()")
	fmt.Println(m.Data())
}

func def() {
	fmt.Println("call def() ............")
}

func TestReflectFunc(t *testing.T) {
	callAbcFunc()
	fmt.Println("--------------------------")
	callDefFunc()
}

func callAbcFunc() {

	fn := abc
	t := reflect.TypeOf(fn)
	fmt.Println(t)
	v := reflect.ValueOf(fn)
	fmt.Println(v)

	var paramsValue []reflect.Value
	paramsValue = append(paramsValue, reflect.ValueOf(&userSession{}))
	paramsValue = append(paramsValue, reflect.ValueOf(&userMessage{}))

	for i := 0; i < 100; i++ {
		runFunc(t, v, paramsValue)
	}
}

func callDefFunc() {
	fn := def
	t := reflect.TypeOf(fn)
	fmt.Println(t)
	v := reflect.ValueOf(fn)
	fmt.Println(v)

	runFunc(t, v, nil)
}

func runFunc(t reflect.Type, v reflect.Value, args []reflect.Value) {
	if t.Kind() != reflect.Func {
		panic("fn is not function.")
	}

	numIn := t.NumIn()
	fmt.Println(numIn)

	if numIn != len(args) {
		panic("fn parameter num mismatch")
	}

	//var params []reflect.Type
	//for c := 0; c < intNum; c++ {
	//	params = append(params, t.In(c))
	//}
	//for _, param := range params {
	//	fmt.Println(param)
	//}

	v.Call(args)

}
