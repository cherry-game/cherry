package test_goroutine

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"
)

type TaskFunc func(ctx context.Context)

func runTaskFunc(wg *sync.WaitGroup, ctx context.Context, taskName string, f TaskFunc) {
	defer wg.Done()

	log.Printf("Task %s start!\n", taskName)
	f(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Task %s revice exit signal \n", taskName)
			return
		}
	}
}

func TestStop2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ctxb, cancelb := context.WithCancel(ctx)
	ctxe, cancele := context.WithCancel(ctx)

	wg := sync.WaitGroup{}

	wg.Add(1)

	go runTaskFunc(&wg, ctx, "A", func(ctx context.Context) {
		wg.Add(1)

		go runTaskFunc(&wg, ctxb, "b", func(ctx context.Context) {
			wg.Add(1)

			go runTaskFunc(&wg, ctx, "C", func(ctx context.Context) {
				wg.Add(1)

				go runTaskFunc(&wg, ctx, "D", func(ctx context.Context) {
				})
			})
		})
	})

	wg.Add(1)

	go runTaskFunc(&wg, ctxe, "E", func(ctx context.Context) {
		wg.Add(1)

		go runTaskFunc(&wg, ctx, "F", func(ctx context.Context) {
			wg.Add(1)

			go runTaskFunc(&wg, ctx, "G", func(ctx context.Context) {

			})
		})
	})

	go func() {
		time.Sleep(3 * time.Second)
		cancele()

		time.Sleep(3 * time.Second)
		cancelb()

		time.Sleep(3 * time.Second)
		cancel()
	}()

	wg.Wait()
}
