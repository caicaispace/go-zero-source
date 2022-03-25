package limit_test

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gozerosource/code/limit/core/limit"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

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
