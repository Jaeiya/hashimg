package internal

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	fPath "path/filepath"
	"strings"
	"sync"

	"github.com/jaeiya/hashimg/internal/utils"
)

type HashResult struct {
	newHashesInfo []HashInfo
	oldHashesInfo map[string]HashInfo
}

type HashInfo struct {
	isNovel bool
	hash    string
	path    string
	cached  bool
	err     error
}

type HasherConfig struct {
	// The smaller this is, the higher chance of collisions
	Length int
	// How many goroutines should be in pool
	Threads int
	// Queue channel minimum size
	QueueSize int
	// Container for hash results
	HashResult *HashResult
	// Should be a unique string
	Prefix     string
	BufferSize int64
}

type Hasher struct {
	mux        sync.Mutex
	threadPool *utils.ThreadPool
	cfg        HasherConfig
}

/*
NewHasher creates a Hasher which takes a HasherConfig.

🟡 The HashConfig.Prefix should be a unique string, because it identifies
which images have been renamed.
*/
func NewHasher(cfg HasherConfig) (*Hasher, error) {
	if len(cfg.Prefix) < 3 {
		return nil, ErrHashPrefixTooShort
	}

	if cfg.HashResult == nil {
		return nil, ErrHashInfoNil
	}

	if cfg.HashResult.oldHashesInfo == nil {
		cfg.HashResult.oldHashesInfo = make(map[string]HashInfo)
	}

	if cfg.Length < 10 {
		return nil, ErrHashLengthTooShort
	}

	tp, err := utils.NewThreadPool(cfg.Threads, cfg.QueueSize, false)
	if err != nil {
		return nil, err
	}

	return &Hasher{
		threadPool: tp,
		cfg:        cfg,
	}, nil
}

func (h *Hasher) Hash(
	fileName string,
	cs CacheStatus,
	filePath string,
	callBack func(cs CacheStatus),
) {
	h.threadPool.Queue(func() {
		hi := HashInfo{path: filePath}

		if cs == Cached {
			ext := fPath.Ext(fileName)
			hi.hash = strings.TrimPrefix(fileName[0:len(fileName)-len(ext)], h.cfg.Prefix)
			hi.cached = true
		} else {
			hi.hash, hi.err = h.computeHash(filePath)
		}

		h.mux.Lock()
		if cs == Cached {
			h.cfg.HashResult.oldHashesInfo[hi.hash] = hi
		} else {
			h.cfg.HashResult.newHashesInfo = append(h.cfg.HashResult.newHashesInfo, hi)
		}
		h.mux.Unlock()
		callBack(cs)
	})
}

func (h *Hasher) Wait() {
	h.threadPool.Wait()
}

func (h *Hasher) computeHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var buf *bufio.Reader
	if h.cfg.BufferSize > 0 {
		buf = bufio.NewReaderSize(file, int(h.cfg.BufferSize))
	} else {
		buf = bufio.NewReader(file)
	}

	sha := sha256.New()
	if _, err := io.Copy(sha, buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha.Sum(nil))[0:h.cfg.Length], nil
}
