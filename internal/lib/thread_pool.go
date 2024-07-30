package lib

import (
	"sync"
)

type ThreadPool[T any] struct {
	wg         sync.WaitGroup
	queue      chan func() T
	ResultChan chan T
	sendResult bool
}

func NewThreadPool[T any](threadCount int, queueSize int, isUsingResult bool) *ThreadPool[T] {
	if queueSize < 10 {
		panic("queue size should be at least 10")
	}

	if threadCount < 2 {
		panic("thread count must be at least 2")
	}

	tp := &ThreadPool[T]{
		queue:      make(chan func() T, queueSize),
		ResultChan: make(chan T, queueSize),
	}
	tp.wg.Add(threadCount)

	tp.sendResult = isUsingResult

	for range threadCount {
		go tp.threadWorker()
	}

	return tp
}

func (tp *ThreadPool[T]) Queue(work func() T) {
	tp.queue <- work
}

func (tp *ThreadPool[T]) QueueNoReturn(work func()) {
	tp.queue <- func() T {
		work()
		var zero T
		return zero
	}
}

func (tp *ThreadPool[T]) Wait() {
	close(tp.queue) // Prevents hanging forever
	tp.wg.Wait()
	if tp.sendResult {
		close(tp.ResultChan)
	}
}

func (tp *ThreadPool[T]) threadWorker() {
	defer tp.wg.Done()
	for work := range tp.queue {
		if tp.sendResult {
			tp.ResultChan <- work()
		}
		work()
	}
}
