package shedding_test

import (
	"fmt"
	"testing"
	"time"

	"gozerosource/code/shedding/core/load"
	"gozerosource/code/shedding/core/stat"
)

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
