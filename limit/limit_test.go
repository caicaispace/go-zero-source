package limit_test

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gozerosource/limit/core/limit"
	"gozerosource/limit/core/syncx"

	"github.com/zeromicro/go-zero/core/stores/redis"
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

	latch := syncx.NewLimit(10)

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

func Test_TokenLimiter(t *testing.T) {
	const (
		burst   = 100
		rate    = 100
		seconds = 5
	)
	store := redis.New("127.0.0.1:6379")
	fmt.Println(store.Ping())
	// New tokenLimiter
	limiter := limit.NewTokenLimiter(rate, burst, store, "token-limiter")
	timer := time.NewTimer(time.Second * seconds)
	quit := make(chan struct{})
	defer timer.Stop()
	go func() {
		<-timer.C
		close(quit)
	}()

	var allowed, denied int32
	var wait sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wait.Add(1)
		go func() {
			for {
				select {
				case <-quit:
					wait.Done()
					return
				default:
					if limiter.Allow() {
						atomic.AddInt32(&allowed, 1)
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

func Test_PeriodLimit(t *testing.T) {
	const (
		burst   = 100
		rate    = 100
		seconds = 5
		threads = 2
	)
	store := redis.New("127.0.0.1:6379")
	fmt.Println(store.Ping())
	lmt := limit.NewPeriodLimit(seconds, 5, store, "period-limit")
	timer := time.NewTimer(time.Second * seconds)
	quit := make(chan struct{})
	defer timer.Stop()
	go func() {
		<-timer.C
		close(quit)
	}()

	var allowed, denied int32
	var wait sync.WaitGroup
	for i := 0; i < threads; i++ {
		i := i
		wait.Add(1)
		go func() {
			for {
				select {
				case <-quit:
					wait.Done()
					return
				default:
					if v, err := lmt.Take(strconv.FormatInt(int64(i), 10)); err == nil && v == limit.Allowed {
						atomic.AddInt32(&allowed, 1)
					} else if err != nil {
						log.Fatal(err)
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
