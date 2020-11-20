package main

import (
	"fmt"
	"time"
)

var c chan int

func main() {
	c = make(chan int, 100)

	for i := 0; i < 10; i++ {
		c <- i
	}

	for i := 0; i <= 2; i++ {
		go process(i)
	}

	time.Sleep(time.Second * 60)
}

func process(index int) {
	for {
		select {
		case i := <-c:
			{
				fmt.Println(fmt.Sprintf("index=%d , value=%d", index, i))
			}
		}
	}
}
