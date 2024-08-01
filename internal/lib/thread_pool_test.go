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
		a.PanicsWithValue("queue size should be at least 10", func() {
			NewThreadPool(2, 5, true)
		})
	})

	t.Run("should error if thread count is too small", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		a.PanicsWithValue("thread count must be at least 2", func() {
			NewThreadPool(1, 10, true)
		})
	})

	t.Run("should queue work", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		tp := NewThreadPool(5, 100, true)

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
