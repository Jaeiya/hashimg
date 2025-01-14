package models

import (
	"sync/atomic"
	"time"
)

type ProcessStatus struct {
	TotalImageCount  int32
	DupeImageCount   int32
	CachedImageCount int32
	NewImageCount    int32
	BufferSize       int64
	// Progress of calculating file hashes
	HashProgress int32
	// Progress of renaming and/or removing files
	UpdateProgress int32
	// The total amount of image hashes that are to be processed
	MaxHashProgress int32
	// The total amount of renames and removals to be processed
	MaxUpdateProgress  int32
	ProcessingComplete bool
	UpdatingComplete   bool
	HashingTook        time.Duration
	UpdatingTook       time.Duration
	FilterTook         time.Duration
	AnalyzeTook        time.Duration
	TotalTime          time.Duration
	HashErr            error
	UpdateErr          error
}

// IncProgress atomically increments HashProgress. This makes it
// thread-safe.
func (ps *ProcessStatus) IncHashProgress() {
	atomic.AddInt32(&ps.HashProgress, 1)
}

// IncUpdateProgress atomically increments UpdateProgress. This makes it
// thread-safe.
func (ps *ProcessStatus) IncUpdateProgress() {
	atomic.AddInt32(&ps.UpdateProgress, 1)
}

// IncCachedImages atomically increments CachedImages. This makes it
// thread-safe.
func (ps *ProcessStatus) IncCachedImages() {
	atomic.AddInt32(&ps.CachedImageCount, 1)
}
