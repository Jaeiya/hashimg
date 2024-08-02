package lib

import (
	"fmt"
	"os"
	fPath "path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jaeiya/go-template/internal/lib/utils"
)

type ProcessStats struct {
	Total int
	New   int
	Dup   int
}

type ImageProcessor struct {
	hashPrefix string
}

func NewImageProcessor(hashPrefix string) ImageProcessor {
	return ImageProcessor{hashPrefix}
}

func (ip ImageProcessor) Process(dir string, hashLen int, iMap ImageMap) (ProcessStats, error) {
	start := time.Now()
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
		Threads:   runtime.NumCPU(),
		QueueSize: queueSize,
		HashInfo:  &hi,
		Prefix:    ip.hashPrefix,
	})
	if err != nil {
		return ProcessStats{}, err
	}

	for fileName, cacheStatus := range iMap {
		hasher.Hash(fileName, cacheStatus, fPath.Join(dir, fileName))
	}

	hasher.Wait()
	fmt.Println("HashSpeed:", time.Since(start))
	fi := &FilteredImages{}
	imgFilter := NewImageFilter()
	imgFilter.FilterImages(hi, fi)
	return ip.updateImages(*fi)
}

func (ip ImageProcessor) updateImages(fi FilteredImages) (ProcessStats, error) {
	start := time.Now()
	if len(fi.dupeImageHashes) == 0 && len(fi.newImageHashes) == 0 {
		return ProcessStats{}, nil
	}

	workLen := len(fi.dupeImageHashes) + len(fi.newImageHashes)
	queueSize := workLen
	if queueSize < 10 {
		queueSize = 10
	}

	tp, err := utils.NewThreadPool(runtime.NumCPU(), queueSize, false)
	if err != nil {
		return ProcessStats{}, err
	}

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
			newFileName := fPath.Join(dir, ip.hashPrefix+newImgHash+ext)
			err := os.Rename(imgPath, newFileName)
			if err != nil {
				mux.Lock()
				errors = append(errors, err)
				mux.Unlock()
			}
		})
	}

	tp.Wait()
	fmt.Println("UpdateSpeed:", time.Since(start))
	if len(errors) > 0 {
		return ProcessStats{}, fmt.Errorf("update errors: %v", errors)
	}

	return ProcessStats{
		Total: workLen,
		New:   len(fi.newImageHashes),
		Dup:   len(fi.dupeImageHashes),
	}, nil
}
