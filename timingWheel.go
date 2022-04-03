package timi

import (
	"fmt"
	"time"
)

type TimingWheel struct {
	now       int64
	tickMs    int64
	wheelSize int64
	slots     [][]job
	next      *TimingWheel
	index     int64
	accept    chan *Job
}

type Job struct {
	job       func()
	timestamp int64
}

type job func()

func timestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func newTimingWheel(tickMs, wheelSize int64) *TimingWheel {
	return &TimingWheel{
		tickMs:    tickMs,
		wheelSize: wheelSize,
		next:      nil,
		slots:     make([][]job, wheelSize),
		accept:    make(chan *Job),
	}
}

func (t *TimingWheel) makeNext() *TimingWheel {
	return &TimingWheel{
		now:       t.now,
		tickMs:    t.tickMs * t.wheelSize,
		wheelSize: t.wheelSize,
		slots:     make([][]job, t.wheelSize),
		next:      nil,
		index:     0,
	}
}

func (t *TimingWheel) add(job func(), after int64) {
	margin := t.tickMs * (t.wheelSize - t.index)
	if after < margin {
		fmt.Printf("now is %d, gap is %d\n", t.now, after/t.tickMs)
		idx := t.index + after/t.tickMs
		t.slots[idx] = append(t.slots[idx], job)
	} else {
		if t.next == nil {
			t.next = t.makeNext()
		}
		t.next.add(job, after-margin)
	}
}

func (t *TimingWheel) Start() {
	t.now = timestamp()
	fmt.Printf("timing wheel start time is %d\n", t.now)
	go func() {
		updated := true
		var timer <-chan time.Time
		for {
			if updated {
				timer = time.After(time.Millisecond * time.Duration(t.tickMs))
				updated = false
			}
			select {
			case heheda := <-timer:
				fmt.Printf("tick time is is %v , tick time ms is %v, "+
					"timestamp is %v\n",
					heheda,
					heheda.UnixMilli(), timestamp())
				jobs := t.slots[t.index]
				for _, job := range jobs {
					go job()
				}
				t.slots[t.index] = make([]job, 0, 0)
				t.index = (t.index + 1) % t.wheelSize
				t.now = t.now + t.tickMs
				updated = true
			case job := <-t.accept:
				fmt.Printf("accept job %v, gap is %d \n", *job,
					job.timestamp-t.now)
				t.add(job.job, job.timestamp-t.now)
			}
		}
	}()
}

func (t *TimingWheel) Offer(job *Job) bool {
	// todo1 get now8
	if job.timestamp <= t.now {
		return false
	}
	select {
	case t.accept <- job:
		return true
	}
}
