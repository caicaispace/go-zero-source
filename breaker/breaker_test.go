package breaker

import (
	"fmt"
	"testing"
	"time"

	"gozerosource/breaker/core/breaker"
)

// https://juejin.cn/post/6891836358155829262

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
				fmt.Println("err", err)
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
