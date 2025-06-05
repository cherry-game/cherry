package cherryTimeWheel

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestAddFunc(t *testing.T) {
	tw := NewTimeWheel(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	id1 := NextID()
	tw.AfterFunc(id1, time.Second, func() {
		fmt.Println("The timer fires")
	})
	fmt.Println(id1)

	id2 := NextID()
	tw.AddEveryFunc(id2, 500*time.Millisecond, func() {
		log.Println("500 Millisecond")
	})

	fmt.Println(id2)

	//for i := 0; i < 10000; i++ {
	//	tw.BuildEveryFunc(10*time.Millisecond, func() {
	//		log.Println("10 Millisecond")
	//	})
	//}

	time.Sleep(time.Hour)
}
