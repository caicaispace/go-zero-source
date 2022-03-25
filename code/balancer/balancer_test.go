package balancer_test

import (
	"testing"
	"time"

	"gozerosource/code/balancer"
)

func TestP2C(t *testing.T) {
	const (
		seconds = 10
	)
	timer := time.NewTimer(time.Second * seconds)
	quit := make(chan struct{})

	defer timer.Stop()
	go func() {
		<-timer.C
		close(quit)
	}()

	go balancer.NewServer()
	go balancer.NewClient()

	for {
		select {
		case <-quit:
			return
		default:
		}
	}
}
