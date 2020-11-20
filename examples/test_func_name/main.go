package main

import (
	"fmt"
	"github.com/phantacix/cherry/utils"
)

type FunStruct struct {
}

func (f *FunStruct) A() string {
	return "A method"
}

func main() {
	f := &FunStruct{}

	result0 := cherryUtils.Reflect.GetFuncName(f)
	fmt.Println(result0)

	result1 := cherryUtils.Reflect.GetFuncName(f.A)
	fmt.Println(result1)

	result2 := cherryUtils.Reflect.GetFuncName(main)
	fmt.Println(result2)

	result3 := cherryUtils.Reflect.GetFuncName(main1)
	fmt.Println(result3)
}

func main1() {

}
