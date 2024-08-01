package lib

import (
	"sync"
)

type ThreadPool struct {
	wg         sync.WaitGroup
	queue      chan func()
	sendResult bool
}

func NewThreadPool(threadCount int, queueSize int, isUsingResult bool) *ThreadPool {
	if queueSize < 10 {
		panic("queue size should be at least 10")
	}

	if threadCount < 2 {
		panic("thread count must be at least 2")
	}

	tp := &ThreadPool{
		queue: make(chan func(), queueSize),
	}
	tp.wg.Add(threadCount)

	tp.sendResult = isUsingResult

	for range threadCount {
		go tp.threadWorker()
	}

	return tp
}

func (tp *ThreadPool) Queue(work func()) {
	tp.queue <- work
}

func (tp *ThreadPool) Wait() {
	close(tp.queue) // Prevents hanging forever
	tp.wg.Wait()
}

func (tp *ThreadPool) threadWorker() {
	defer tp.wg.Done()
	for work := range tp.queue {
		work()
	}
}
