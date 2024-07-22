package lib

import (
	"crypto/sha256"
	"fmt"
	"io"
)

type ImgHasher interface {
	Hash(reader io.Reader, path string, result chan<- HashResult)
}

type MyHasher struct{}

func (MyHasher) Hash(reader io.Reader, path string, result chan<- HashResult) {
	h := sha256.New()

	if _, err := io.Copy(h, reader); err != nil {
		result <- HashResult{
			err: err,
		}
	}

	result <- HashResult{
		hash:   fmt.Sprintf("%x", h.Sum(nil))[0:24],
		path:   path,
		cached: false,
		err:    nil,
	}
}
