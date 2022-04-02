package timi

import (
	"fmt"
	"testing"
	"time"
)

func Test_no(t *testing.T) {
	fmt.Println(timestamp())
}

func Test_timingWheel(t *testing.T) {
	timingWheel := newTimingWheel(int64(time.Second/time.Millisecond), 10)
	timingWheel.Start()
	now := timestamp()
	for i := 0; i < 10; i++ {
		funcTime := now + (int64(i)+10)*1000
		fmt.Println(timingWheel.Offer(&Job{
			job: func() {
				now := timestamp()
				fmt.Printf("prepare time %d, now is %d, "+
					"gap is %d\n", funcTime, now, funcTime-now)
			},
			timestamp: funcTime,
		}))
	}
	select {}
}
