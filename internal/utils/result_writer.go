package utils

import (
	"sync"
)

type ResultRenderer[T any] interface {
	Render(T)
}

type ResultWriter[T any] struct {
	wg      *sync.WaitGroup
	resChan chan T
}

func NewResultWriter[T any](rr ResultRenderer[T]) *ResultWriter[T] {
	resChan := make(chan T, 500)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for res := range resChan {
			rr.Render(res)
		}
	}()

	return &ResultWriter[T]{wg: &wg, resChan: resChan}
}

func (rw *ResultWriter[T]) Write(res T) {
	select {
	case rw.resChan <- res:
	default:
		// Skip writes that are too fast
	}
}

func (rw *ResultWriter[T]) Close() {
	close(rw.resChan)
	rw.wg.Wait()
}
