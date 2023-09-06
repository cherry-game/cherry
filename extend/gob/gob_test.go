package cherryGOB

import (
	"fmt"
	"reflect"
	"testing"

	cproto "github.com/cherry-game/cherry/net/proto"
)

func TestPB(t *testing.T) {
	rsp := &cproto.Response{
		Code: 11,
		Data: []byte{1, 2, 3},
	}

	gobBytes, err := Encode(rsp)
	if err != nil {
		fmt.Println(err)
		return
	}

	rsp1Type := reflect.TypeOf(rsp)

	x, err := Decode(gobBytes, []reflect.Type{rsp1Type})
	fmt.Println(x, err)

	rsp1, ok := x[0].Interface().(*cproto.Response)
	fmt.Println(rsp1, ok)

}

func TestCallFunc(t *testing.T) {
	type T1 struct {
		A int
		B string
		C int32
	}

	var (
		a  = 1
		b  = 2
		t1 = &T1{A: 1, B: "2", C: 3}
	)

	gobBytes, err := Encode(a, b, t1)
	if err != nil {
		fmt.Println(err)
		return
	}

	fn := func(a int, b int, c *T1) {
		fmt.Println("ok!!!!!!!", a, b, c)
	}

	paramsType := reflect.TypeOf(fn)
	paramsValue := reflect.ValueOf(fn)

	decodeValue, err := DecodeFunc(gobBytes, paramsType)
	if err != nil {
		fmt.Println(err)
		return
	}

	paramsValue.Call(decodeValue)
}
