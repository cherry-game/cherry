package main

import (
	"fmt"
	"github.com/cherry-game/cherry/extend/reflect"
)

type FunStruct struct {
}

func (f *FunStruct) A() string {
	return "A method"
}

func main() {
	f := &FunStruct{}

	result0 := cherryReflect.GetFuncName(f)
	fmt.Println(result0)

	result1 := cherryReflect.GetFuncName(f.A)
	fmt.Println(result1)

	result2 := cherryReflect.GetFuncName(main)
	fmt.Println(result2)

	result3 := cherryReflect.GetFuncName(main1)
	fmt.Println(result3)
}

func main1() {

}
