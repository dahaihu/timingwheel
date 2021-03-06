package timi

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type TimingWheel struct {
	now       time.Time
	tickMs    int64
	wheelSize int64
	slots     [][]*Job
	index     int64
	accept    chan *Job
	next      *TimingWheel
	log       *zap.Logger
}

type Job struct {
	job       func()
	runtime   int64
	addedTime time.Time
	added     bool
	after     int64
	wg        *sync.WaitGroup
}

func (t *TimingWheel) timestamp() int64 {
	return t.now.UnixMilli()
}

func (t *TimingWheel) newJob(runtime int64, job func()) *Job {
	newJob := &Job{
		job:     job,
		runtime: runtime,
		wg:      new(sync.WaitGroup),
	}
	newJob.wg.Add(1)
	return newJob
}

func (j *Job) run(logger *zap.Logger, tickTime int64) {
	logger.Debug("job run",
		zap.Int64("runtime", j.runtime),
		zap.Int64("added_time", j.addedTime.UnixMilli()),
		zap.Int64("now", tickTime),
		zap.Int64("real_gap", j.runtime-tickTime),
		zap.Int64("added_gap", tickTime-j.addedTime.UnixMilli()))
	j.job()
}

func timestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func newTimingWheel(tickMs, wheelSize int64, log *zap.Logger) *TimingWheel {
	return &TimingWheel{
		tickMs:    tickMs,
		wheelSize: wheelSize,
		next:      nil,
		slots:     make([][]*Job, wheelSize),
		accept:    make(chan *Job),
		log:       log,
	}
}

func (t *TimingWheel) makeNext() *TimingWheel {
	return &TimingWheel{
		now:       t.now,
		tickMs:    t.tickMs * t.wheelSize,
		wheelSize: t.wheelSize,
		slots:     make([][]*Job, t.wheelSize),
		next:      nil,
		index:     0,
		log:       t.log,
	}
}

func (t *TimingWheel) add(job *Job, after int64, level int) {
	margin := t.tickMs * (t.wheelSize - t.index)
	if after < margin {
		idx := t.index + after/t.tickMs
		job.after = after % t.tickMs
		t.log.Debug("job added",
			zap.Int("level", level),
			zap.Int64("index", idx),
			zap.Int64("gap", after),
			zap.Int64("after", job.after),
			zap.Int64("runtime", job.runtime),
			zap.Int64("now", t.timestamp()))
		t.slots[idx] = append(t.slots[idx], job)
	} else {
		if t.next == nil {
			t.next = t.makeNext()
		}
		t.next.add(job, after-margin, level+1)
	}
}

func (t *TimingWheel) advance() []*Job {
	jobs := t.slots[t.index]
	t.slots[t.index] = make([]*Job, 0)
	t.index = t.index + 1
	t.now = t.now.Add(time.Duration(t.tickMs) * time.Millisecond)
	if t.index == t.wheelSize {
		t.index = 0
		if t.next != nil {
			nextJobs := t.next.advance()
			for _, job := range nextJobs {
				t.add(job, job.after, 0)
			}
		}
	}
	return jobs
}

func (t *TimingWheel) tickDuration() time.Duration {
	return t.now.Add(time.Duration(t.tickMs) * time.Millisecond).Sub(time.Now())
}

func (t *TimingWheel) Start() {
	t.now = time.Now()
	go func() {
		ticker := time.NewTicker(t.tickDuration())
		for {
			select {
			case tickTime := <-ticker.C:
				t.log.Debug("tick time",
					zap.Int64("now", tickTime.UnixMilli()),
					zap.Int64("cur", timestamp()),
					zap.Int64("timingWheel now", t.timestamp()),
					zap.Int64("index", t.index))
				jobs := t.advance()
				for _, job := range jobs {
					go job.run(t.log, tickTime.UnixMilli())
				}
				ticker.Reset(t.tickDuration())
			case job := <-t.accept:
				if job.runtime < t.timestamp() {
					job.added = false
				} else {
					job.added = true
					job.addedTime = t.now
					t.add(job, job.runtime-t.timestamp(), 0)
				}
				job.wg.Done()
			}
		}
	}()
}

func (t *TimingWheel) Offer(runtime int64, job func()) bool {
	transferJob := t.newJob(runtime, job)
	t.accept <- transferJob
	transferJob.wg.Wait()
	return transferJob.added
}
