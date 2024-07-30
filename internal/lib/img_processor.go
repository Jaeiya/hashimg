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

func ProcessImages(dir string, iMap ImageMap) (ProcessStats, error) {
	mapLen := len(iMap)
	if mapLen == 0 {
		return ProcessStats{}, fmt.Errorf("no images found in %s", dir)
	}

	queueSize := mapLen
	if queueSize < 10 {
		queueSize = 10
	}

	hi := []HashInfo{}
	hasher := NewHasher(HasherConfig{
		Length:    24,
		Threads:   10,
		QueueSize: queueSize,
		HashInfo:  &hi,
	})

	for fileName, cacheStatus := range iMap {
		hasher.Hash(fileName, cacheStatus, fPath.Join(dir, fileName))
	}

	hasher.Wait()
	fi := &FilteredImages{}
	imgFilter := NewImageFilter()
	imgFilter.FilterImages(hi, fi)
	return updateImages(*fi)
}

func updateImages(fr FilteredImages) (ProcessStats, error) {
	if len(fr.dupeImageHashes) == 0 && len(fr.newImageHashes) == 0 {
		return ProcessStats{}, nil
	}

	workLen := len(fr.dupeImageHashes) + len(fr.newImageHashes)
	queueSize := workLen
	if queueSize < 10 {
		queueSize = 10
	}

	tp := NewThreadPool[error](10, queueSize, false)

	errors := []error{}
	mux := sync.Mutex{}

	for _, imgPath := range fr.dupeImageHashes {
		tp.QueueNoReturn(func() {
			err := os.Remove(imgPath)
			if err != nil {
				mux.Lock()
				errors = append(errors, err)
				mux.Unlock()
			}
		})
	}

	for newImgHash, imgPath := range fr.newImageHashes {
		tp.QueueNoReturn(func() {
			dir := fPath.Dir(imgPath)
			// Extensions should always be lowercase even though the
			// file system doesn't care
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
		New:   len(fr.newImageHashes),
		Dup:   len(fr.dupeImageHashes),
	}, nil
}
