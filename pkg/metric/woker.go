package metric

import (
	"sync/atomic"

	"github.com/Gonewithmyself/gobot/pkg/mpsc"
)

const (
	workerIdle = iota
	workerWork
)

type Worker struct {
	// ch   chan *Metric
	id       int32
	status   int32
	queue    *mpsc.Queue
	wakeupCh chan struct{}
	donec    chan struct{}
}

func NewWorker(id int32) *Worker {
	return &Worker{
		// ch:   make(chan *Metric),
		id:       id,
		queue:    mpsc.New(),
		wakeupCh: make(chan struct{}),
		donec:    make(chan struct{}),
	}
}

func (worker *Worker) run() {
	defer func() {
		worker.drain()
	}()

	for {
		switch atomic.LoadInt32(&worker.status) {
		case workerIdle:
			select {
			case <-worker.wakeupCh:
			case <-worker.donec:
				return
			}

		case workerWork:
			worker.drain()
			atomic.CompareAndSwapInt32(&worker.status, workerWork, workerIdle)
		}
	}
}

// 一次性消费完队列
func (worker *Worker) drain() {
	for {
		v := worker.queue.Pop()
		if v == nil {
			break
		}

		metric := v.(IMetric)
		Mgr.handler.ProcessMetric(metric, GetRecorder(metric.GetMsgId()))
		atomic.AddInt64(&Mgr.pending, -1)
	}
}

func (worker *Worker) submit(metric IMetric) {
	// worker.ch <- metric
	worker.queue.Push(metric)
	if atomic.LoadInt32(&worker.status) == workerIdle {
		if atomic.CompareAndSwapInt32(&worker.status, workerIdle, workerWork) {
			worker.wakeupCh <- struct{}{}
		}
	}
}
