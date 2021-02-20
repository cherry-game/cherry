package main

import (
	"fmt"
	"gorm.io/gorm/utils"
	"sort"
)

type intType []int

func main() {
	x := &intType{
		40, 21, 62, 33,
	}

	x.Order()
	fmt.Println(x)

	fmt.Print(x.Test())
}

func (s intType) Test() string {
	var result string
	for i := 0; i < len(s); i++ {
		result += utils.ToString(s[i]) + "-"
	}

	return result
}

func (s intType) Order() {
	sort.Ints(s)
}
