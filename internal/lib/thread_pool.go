package lib

import (
	"sync"
)

type ThreadPool struct {
	wg    sync.WaitGroup
	queue chan func()
}

func NewThreadPool(threadCount int, queueSize int) *ThreadPool {
	tp := &ThreadPool{
		queue: make(chan func(), queueSize),
	}
	tp.wg.Add(threadCount)

	for range threadCount {
		go tp.threadWorker()
	}

	return tp
}

func (tp *ThreadPool) Queue(work func()) error {
	tp.queue <- work
	return nil
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
