package timi

import (
	"fmt"
	"math/rand"
	"testing"

	"go.uber.org/zap"
)

func Test_no(t *testing.T) {
	fmt.Println(timestamp())
}

func Test_timingWheel(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	timingWheel := newTimingWheel(100, 10, logger)
	timingWheel.Start()
	now := timestamp()
	for i := 0; i < 10; i++ {
		funcTime := now + 1000 + rand.Int63n(10) - 5
		timingWheel.Offer(funcTime, func() {
			fmt.Println("hello world")
		})
	}
	select {}
}
