package lib

import (
	"io"
	"os"
)

type ImgFileOpener interface {
	Open(path string) (io.ReadCloser, error)
}

type MyFileOpener struct{}

func (MyFileOpener) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}
