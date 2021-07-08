package main

import (
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"time"
)

func defaultFor(ids []int) {
	startTime := time.Now().UnixNano() / 1e6
	for _, id := range ids {
		fmt.Println(id)
	}
	endTime := time.Now().UnixNano() / 1e6
	fmt.Println("defaultFor总耗时：", endTime-startTime)
}

func linqFor(ids []int) {
	startTime := time.Now().UnixNano() / 1e6
	linq.From(ids).ForEachT(func(id int) {
		fmt.Println(id)
	})
	endTime := time.Now().UnixNano() / 1e6
	fmt.Println("linqFor总耗时：", endTime-startTime)
}

//-------------------------
