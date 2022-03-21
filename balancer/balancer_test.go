package balancer_test

import (
	"fmt"
	"testing"
	"time"

	"gozerosource/balancer"
)

func TestP2C(t *testing.T) {
	const (
		seconds = 5
	)
	timer := time.NewTimer(time.Second * seconds)
	quit := make(chan struct{})

	defer timer.Stop()
	go func() {
		<-timer.C
		close(quit)
	}()

	go balancer.NewServer()

	select {
	case <-quit:
		fmt.Println("quit >>>>>>>>>>>>>>>>>")
		return
	default:
		balancer.NewClient()
	}
}
