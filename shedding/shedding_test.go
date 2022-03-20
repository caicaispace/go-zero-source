package shedding_test

import (
	"fmt"
	"testing"
	"time"

	"gozerosource/shedding/core/load"
	"gozerosource/shedding/core/stat"
)

func Test_Shedding(t *testing.T) {
	shedder := load.NewAdaptiveShedder(load.WithCpuThreshold(6))
	// cpuFull()
	for i := 0; i < 100; i++ {
		fmt.Println(i, stat.CpuUsage())
		promise, err := shedder.Allow()
		if err != nil {
			fmt.Println(">>>>>>>>>>>>>")
			return
		}
		promise.Fail()
		// promise.Pass()
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(1 * time.Second)
}

func cpuFull() {
	for i := 0; i < 100; i++ {
		go func(i int) {
			for {
			}
		}(i)
	}
}
