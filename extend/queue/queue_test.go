package cherryQueue

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	num := 3

	q := NewQueue()
	for i := 0; i < num; i++ {
		q.Push(i)
	}

	for i := 0; i < num; i++ {
		v, ok := q.Pop()
		fmt.Println(v, ok)
	}

	fmt.Println(fmt.Sprintf("len = %d", q.Len()))
}

func BenchmarkNewFIFOQueue(b *testing.B) {
	q := NewQueue()
	for i := 0; i < 1000000; i++ {
		q.Push(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q.Push(i)
		q.Pop()
	}
}

func TestQueuePop(t *testing.T) {
	num := 10

	q := NewQueue()
	for i := 0; i < num; i++ {
		q.Push(i)
	}

	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			q.Push(rand.Int31n(10000))
		}
	}()

	go func() {
		postTicker := time.NewTicker(5 * time.Millisecond)
		postNum := 100

		for {
			select {
			case <-postTicker.C:
				{
					for i := 0; i < postNum; i++ {
						v, ok := q.Pop()
						if !ok {
							break
						}
						fmt.Println(v)
					}
				}
			}
		}
	}()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
