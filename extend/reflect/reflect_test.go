package cherryReflect

import (
	"fmt"
	"testing"
)

type FunStruct struct {
}

func (f *FunStruct) A() string {
	return "A method"
}

func test1111() {}

func TestGetFuncName(t *testing.T) {
	f := &FunStruct{}

	result0 := GetFuncName(f)
	fmt.Println(result0)

	result1 := GetFuncName(f.A)
	fmt.Println(result1)

	result2 := GetFuncName(test1111)
	fmt.Println(result2)
}

func TestIsPtr(t *testing.T) {
	fmt.Println(IsPtr(nil))
}

func TestIsNotPtr(t *testing.T) {
	s := &FunStruct{}
	fmt.Println(IsNotPtr(s))
}
