package lib

import (
	"fmt"
	"os"
	fPath "path/filepath"
	"strings"
	"sync"
)

type ProcessStats struct {
	Total int
	New   int
	Dup   int
}

func ProcessImages(dir string, hashLen int, iMap ImageMap) (ProcessStats, error) {
	mapLen := len(iMap)
	if mapLen == 0 {
		return ProcessStats{}, fmt.Errorf("empty image map")
	}

	queueSize := mapLen
	if queueSize < 10 {
		queueSize = 10
	}

	hi := []HashInfo{}

	hasher, err := NewHasher(HasherConfig{
		Length:    hashLen,
		Threads:   10,
		QueueSize: queueSize,
		HashInfo:  &hi,
	})
	if err != nil {
		return ProcessStats{}, err
	}

	for fileName, cacheStatus := range iMap {
		hasher.Hash(fileName, cacheStatus, fPath.Join(dir, fileName))
	}

	hasher.Wait()
	fi := &FilteredImages{}
	imgFilter := NewImageFilter()
	imgFilter.FilterImages(hi, fi)
	return updateImages(*fi)
}

func updateImages(fi FilteredImages) (ProcessStats, error) {
	if len(fi.dupeImageHashes) == 0 && len(fi.newImageHashes) == 0 {
		return ProcessStats{}, nil
	}

	workLen := len(fi.dupeImageHashes) + len(fi.newImageHashes)
	queueSize := workLen
	if queueSize < 10 {
		queueSize = 10
	}

	tp := NewThreadPool(10, queueSize, false)

	errors := []error{}
	mux := sync.Mutex{}

	for _, imgPath := range fi.dupeImageHashes {
		tp.Queue(func() {
			err := os.Remove(imgPath)
			if err != nil {
				mux.Lock()
				errors = append(errors, err)
				mux.Unlock()
			}
		})
	}

	for newImgHash, imgPath := range fi.newImageHashes {
		tp.Queue(func() {
			dir := fPath.Dir(imgPath)
			// Uppercase extensions are ugly and inconsistent
			ext := strings.ToLower(fPath.Ext(imgPath))
			newFileName := fPath.Join(dir, hashPrefix+newImgHash+ext)
			err := os.Rename(imgPath, newFileName)
			if err != nil {
				mux.Lock()
				errors = append(errors, err)
				mux.Unlock()
			}
		})
	}

	tp.Wait()

	if len(errors) > 0 {
		return ProcessStats{}, fmt.Errorf("update errors: %v", errors)
	}

	return ProcessStats{
		Total: workLen,
		New:   len(fi.newImageHashes),
		Dup:   len(fi.dupeImageHashes),
	}, nil
}
