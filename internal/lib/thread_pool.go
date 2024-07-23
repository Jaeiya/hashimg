package lib

import (
	"sync"
)

type ThreadPool[T any] struct {
	wg         sync.WaitGroup
	queue      chan func() T
	ResultChan chan T
}

func NewThreadPool[T any](threadCount int, queueSize int, resultChan chan T) *ThreadPool[T] {
	if queueSize < 10 {
		panic("queue size should be at least 10")
	}

	if threadCount < 2 {
		panic("thread count must be at least 2")
	}

	tp := &ThreadPool[T]{
		queue:      make(chan func() T, queueSize),
		ResultChan: resultChan,
	}
	tp.wg.Add(threadCount)

	for range threadCount {
		go tp.threadWorker()
	}

	return tp
}

func (tp *ThreadPool[T]) Queue(work func() T) error {
	tp.queue <- work
	return nil
}

func (tp *ThreadPool[T]) Wait() {
	close(tp.queue) // Prevents hanging forever
	tp.wg.Wait()
	close(tp.ResultChan)
}

func (tp *ThreadPool[T]) threadWorker() {
	defer tp.wg.Done()
	for work := range tp.queue {
		tp.ResultChan <- work()
	}
}
