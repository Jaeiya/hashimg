package lib

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	fp "path/filepath"
	"strings"
	"sync"
)

const hashPrefix = "0x@"

type HashInfo struct {
	hash string
	// Path to the file
	path string
	// If image has already been hashed and prefixed
	cached bool
	// If an error occurs during processing
	err error
}

type HasherConfig struct {
	Length    int
	Threads   int
	QueueSize int
	HashInfo  *[]HashInfo
}

func NewHasher(c HasherConfig) *Hasher {
	return &Hasher{
		tp:       NewThreadPool[HashInfo](c.Threads, c.QueueSize, false),
		hashLen:  c.Length,
		hashInfo: c.HashInfo,
	}
}

type Hasher struct {
	mux      sync.Mutex
	tp       *ThreadPool[HashInfo]
	hashLen  int
	hashInfo *[]HashInfo
}

func (h *Hasher) Hash(fileName string, filePath string) {
	h.tp.QueueNoReturn(func() {
		hi := HashInfo{}

		defer func() {
			h.mux.Lock()
			*h.hashInfo = append(*h.hashInfo, hi)
			h.mux.Unlock()
		}()

		if strings.Contains(fileName, hashPrefix) {
			ext := fp.Ext(fileName)
			hi = HashInfo{
				hash:   strings.TrimPrefix(fileName[0:len(fileName)-len(ext)], hashPrefix),
				path:   filePath,
				cached: true,
				err:    nil,
			}
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			hi = HashInfo{
				path: filePath,
				err:  err,
			}
			return
		}
		defer file.Close()

		sha := sha256.New()

		if _, err := io.Copy(sha, file); err != nil {
			hi = HashInfo{
				path: filePath,
				err:  err,
			}
			return
		}

		hi = HashInfo{
			hash:   fmt.Sprintf("%x", sha.Sum(nil))[0:h.hashLen],
			path:   filePath,
			cached: false,
			err:    nil,
		}
	})
}

func (h *Hasher) Wait() {
	h.tp.Wait()
}
