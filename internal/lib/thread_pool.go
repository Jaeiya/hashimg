package lib

import (
	"errors"
	"sync"
)

var (
	ErrThreadPoolQueueTooSmall = errors.New("queue size should be at least 10")
	ErrThreadPoolCountTooSmall = errors.New("thread count must be at least 2")
)

type ThreadPool struct {
	wg         sync.WaitGroup
	queue      chan func()
	sendResult bool
}

func NewThreadPool(threadCount int, queueSize int, isUsingResult bool) (*ThreadPool, error) {
	if queueSize < 10 {
		return nil, ErrThreadPoolQueueTooSmall
	}

	if threadCount < 2 {
		return nil, ErrThreadPoolCountTooSmall
	}

	tp := &ThreadPool{
		queue: make(chan func(), queueSize),
	}
	tp.wg.Add(threadCount)

	tp.sendResult = isUsingResult

	for range threadCount {
		go tp.threadWorker()
	}

	return tp, nil
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
