package limit_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gozerosource/code/limit/core/syncx"
)

func Test_limit(t *testing.T) {
	const (
		seconds = 5
		threads = 100
	)
	timer := time.NewTimer(time.Second * seconds)
	quit := make(chan struct{})
	defer timer.Stop()
	go func() {
		<-timer.C
		close(quit)
	}()

	latch := syncx.NewLimit(20)

	var allowed, denied int32
	var wait sync.WaitGroup
	for i := 0; i < threads; i++ {
		wait.Add(1)
		go func() {
			for {
				select {
				case <-quit:
					wait.Done()
					return
				default:
					if latch.TryBorrow() {
						atomic.AddInt32(&allowed, 1)
						defer func() {
							if err := latch.Return(); err != nil {
								fmt.Println(err)
							}
						}()
					} else {
						atomic.AddInt32(&denied, 1)
					}
				}
			}
		}()
	}

	wait.Wait()
	fmt.Printf("allowed: %d, denied: %d, qps: %d\n", allowed, denied, (allowed+denied)/seconds)
}

func Benchmark_limit(b *testing.B) {
	// 测试一个对象或者函数在多线程的场景下面是否安全
	b.RunParallel(func(pb *testing.PB) {
		latch := syncx.NewLimit(10)
		for pb.Next() {
			if latch.TryBorrow() {
				defer func() {
					if err := latch.Return(); err != nil {
						fmt.Println(err)
					}
				}()
			}
		}
	})
}
