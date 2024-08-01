package lib

import (
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThreadPool(t *testing.T) {
	t.Run("should error if queue is too small", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		_, err := NewThreadPool(2, 5, true)
		a.ErrorIs(err, ErrThreadPoolQueueTooSmall)
	})

	t.Run("should error if thread count is too small", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		_, err := NewThreadPool(1, 10, true)
		a.ErrorIs(err, ErrThreadPoolCountTooSmall)
	})

	t.Run("should queue work", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		tp, err := NewThreadPool(5, 100, true)
		a.NoError(err)

		workCount := 5000

		ints := make([]int, workCount)
		mux := sync.Mutex{}

		for i := 0; i < workCount; i++ {
			tp.Queue(func() {
				mux.Lock()
				ints[i] = i*10 + i*100
				mux.Unlock()
			})
		}

		tp.Wait()
		sort.Ints(ints[:])
		for i, ii := range ints {
			a.Equal(i*10+i*100, ii)
		}
		a.Equal(workCount, len(ints))
	})
}
