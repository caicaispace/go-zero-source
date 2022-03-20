package breaker

import (
	"fmt"
	"testing"
	"time"
)

func TestGoogleBreaker(t *testing.T) {
	// 简单场景直接判断对象是否被熔断，执行请求后必须需手动上报执行结果至熔断器。
	gb := newGoogleBreaker()
	for i := 0; i < 100; i++ {
		allow, err := gb.allow()
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
	fmt.Println(gb.history())

	// 复杂场景下支持自定义快速失败，自定义判定请求是否成功的熔断方法，自动上报执行结果至熔断器。
	gb2 := newGoogleBreaker()
	for i := 0; i < 100; i++ {
		err := gb2.doReq(
			func() error {
				if i < 10 {
					time.Sleep(20 * time.Millisecond)
					// return errors.New(">>>>>>>>>")
				}
				return nil
			},
			func(err error) error {
				fmt.Println("err", err)
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
	fmt.Println(gb2.history())
}
