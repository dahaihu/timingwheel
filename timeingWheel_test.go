package timi

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func Test_no(t *testing.T) {
	fmt.Println(timestamp())
}

func newJob(prepareTime int64) *Job {
	return &Job{
		job: func() {
			now := timestamp()
			fmt.Printf("prepare time is %d, now is %d, gap is %d\n",
				prepareTime,
				now,
				prepareTime-now)
		},
		timestamp: prepareTime,
	}
}

func Test_timingWheel(t *testing.T) {
	timingWheel := newTimingWheel(int64(time.Second/time.Millisecond), 100)
	timingWheel.Start()
	now := timestamp()
	for i := 0; i < 10; i++ {
		funcTime := now + int64(i+1)*1000 + rand.Int63n(10) - 5
		fmt.Println(timingWheel.Offer(newJob(funcTime)))
	}
	select {}
}
