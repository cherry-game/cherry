package main

import (
	"fmt"
	"github.com/modern-go/reflect2"
)

type MyStruct struct {
	a int
	b string
}

func main() {
	fmt.Println("main.........")

	t := reflect2.TypeByName("main.MyStruct")
	fmt.Println(t)

	valType := reflect2.TypeOf(1)
	i := 1
	j := 10
	valType.Set(&i, &j)

	//var configPath = ""
	//var profile = ""
	//var nodeId = ""
	//app := cherry.New(configPath, profile, nodeId)

}
