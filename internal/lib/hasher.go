package lib

import (
	"crypto/sha256"
	"fmt"
	"io"
)

type ImgHasher interface {
	Hash(reader io.Reader, path string, length int) HashResult
}

type MyHasher struct{}

func (MyHasher) Hash(reader io.Reader, path string, length int) HashResult {
	h := sha256.New()

	if _, err := io.Copy(h, reader); err != nil {
		return HashResult{
			err: err,
		}
	}

	return HashResult{
		hash:   fmt.Sprintf("%x", h.Sum(nil))[0:length],
		path:   path,
		cached: false,
		err:    nil,
	}
}
