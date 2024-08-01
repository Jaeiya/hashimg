package lib

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	fPath "path/filepath"
	"strings"
	"sync"
)

var (
	ErrHashPrefixTooShort = errors.New("hash prefix must be at least 3 characters")
	ErrHashInfoNil        = errors.New("hash info is nil; it must be initialized")
	ErrHashLengthTooShort = errors.New("hash length must be at least 10 characters")
)

type HashInfo struct {
	hash   string
	path   string
	cached bool
	err    error
}

type HasherConfig struct {
	// The smaller this is, the higher chance of collisions
	Length int
	// How many goroutines should be in pool
	Threads int
	// Queue channel minimum size
	QueueSize int
	// Container for hash results
	HashInfo *[]HashInfo
	// Should be a unique string
	Prefix string
}

type Hasher struct {
	mux        sync.Mutex
	threadPool *ThreadPool
	hashLen    int
	hashInfo   *[]HashInfo
	prefix     string
}

/*
NewHasher creates a Hasher which takes a HasherConfig.

🟡 The HashConfig.Prefix should be a unique string, because it identifies
which images have been renamed.
*/
func NewHasher(c HasherConfig) (*Hasher, error) {
	if len(c.Prefix) < 3 {
		return nil, ErrHashPrefixTooShort
	}

	if c.HashInfo == nil {
		return nil, ErrHashInfoNil
	}

	if c.Length < 10 {
		return nil, ErrHashLengthTooShort
	}

	tp, err := NewThreadPool(c.Threads, c.QueueSize, false)
	if err != nil {
		return nil, err
	}

	return &Hasher{
		threadPool: tp,
		hashLen:    c.Length,
		hashInfo:   c.HashInfo,
		prefix:     c.Prefix,
	}, nil
}

func (h *Hasher) Hash(fileName string, cs CacheStatus, filePath string) {
	h.threadPool.Queue(func() {
		hi := HashInfo{path: filePath}

		if cs == Cached {
			ext := fPath.Ext(fileName)
			hi.hash = strings.TrimPrefix(fileName[0:len(fileName)-len(ext)], h.prefix)
			hi.cached = true
			hi.err = nil
		} else {
			hi.hash, hi.err = h.computeHash(filePath)
		}

		h.mux.Lock()
		*h.hashInfo = append(*h.hashInfo, hi)
		h.mux.Unlock()
	})
}

func (h *Hasher) Wait() {
	h.threadPool.Wait()
}

func (h *Hasher) computeHash(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()
	buf := bufio.NewReader(file)

	sha := sha256.New()
	if _, err := io.Copy(sha, buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha.Sum(nil))[0:h.hashLen], nil
}
