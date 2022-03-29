package shedding_test

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gozerosource/code/core/load"
	"gozerosource/code/core/stat"

	"github.com/zeromicro/go-zero/core/mathx"
)

const (
	buckets        = 10
	bucketDuration = time.Millisecond * 50
)

func TestAdaptiveShedder(t *testing.T) {
	load.DisableLog()
	shedder := load.NewAdaptiveShedder(
		load.WithWindow(bucketDuration),
		load.WithBuckets(buckets),
		load.WithCpuThreshold(100),
	)
	var wg sync.WaitGroup
	var drop int64
	proba := mathx.NewProba()
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 30; i++ {
				promise, err := shedder.Allow()
				if err != nil {
					atomic.AddInt64(&drop, 1)
				} else {
					count := rand.Intn(5)
					time.Sleep(time.Millisecond * time.Duration(count))
					if proba.TrueOnProba(0.01) {
						promise.Fail()
					} else {
						promise.Pass()
					}
				}
			}
		}()
	}
	wg.Wait()
}

func Test_Shedding(t *testing.T) {
	shedder := load.NewAdaptiveShedder(load.WithCpuThreshold(6))
	// cpuFull()
	for i := 0; i < 100; i++ {
		fmt.Println(i, stat.CpuUsage())
		promise, err := shedder.Allow()
		if err != nil {
			fmt.Println(err)
			return
		}
		promise.Fail()
		// promise.Pass()
		time.Sleep(10 * time.Millisecond)
	}
}

func cpuFull() {
	for i := 0; i < 100; i++ {
		go func(i int) {
			for {
			}
		}(i)
	}
}
