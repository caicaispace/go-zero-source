package breaker

import (
	"fmt"
	"testing"
	"time"

	"gozerosource/code/core/breaker"
)

func Test_Breaker(t *testing.T) {
	b := breaker.NewBreaker()
	for i := 0; i < 100; i++ {
		allow, err := b.Allow()
		if err != nil {
			fmt.Println("err", err)
			break
		}
		if i < 10 {
			allow.Reject()
			// time.Sleep(2000 * time.Millisecond)
			time.Sleep(20 * time.Millisecond)
		} else {
			allow.Accept()
		}
	}
	fmt.Println(b.GB.History())
}

func Test_Beaker2(t *testing.T) {
	b := breaker.NewBreaker()
	for i := 0; i < 100; i++ {
		err := b.DoWithAcceptable(
			func() error {
				if i < 10 {
					time.Sleep(20 * time.Millisecond)
					// return errors.New(">>>>>>>>>")
				}
				return nil
			},
			func(err error) bool {
				// fmt.Println("err", err)
				return i >= 8
			},
		)
		if err != nil {
			fmt.Println("err", err)
			break
		}
	}
	fmt.Println(b.GB.History())
}

func Benchmark_BrewkerSerial(b *testing.B) {
	bk := breaker.NewBreaker()
	for i := 0; i < b.N; i++ {
		allow, err := bk.Allow()
		if err != nil {
			fmt.Println("err", err)
			break
		}
		if i%2 == 0 {
			allow.Accept()
		} else {
			allow.Reject()
		}
	}
	fmt.Println(bk.GB.History())
}

func Benchmark_BrewkerParallel(b *testing.B) {
	// 测试一个对象或者函数在多线程的场景下面是否安全
	b.RunParallel(func(pb *testing.PB) {
		bk := breaker.NewBreaker()
		i := 0
		for pb.Next() {
			i++
			allow, err := bk.Allow()
			if err != nil {
				fmt.Println("err", err)
				break
			}
			if i%100 == 0 {
				allow.Accept()
			} else {
				allow.Reject()
			}
			// time.Sleep(20 * time.Millisecond)
		}
		fmt.Println(bk.GB.History())
	})
}
