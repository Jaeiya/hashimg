package lib

import (
	"errors"
	"fmt"
	"os"
	fPath "path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jaeiya/go-template/internal/lib/models"
	"github.com/jaeiya/go-template/internal/lib/utils"
)

var ErrNoImages = errors.New("no images found in directory")

type ImageProcessor struct {
	hashPrefix       string
	imageMap         ImageMap
	processStatus    *models.ProcessStatus
	processStartTime time.Time
}

func NewImageProcessor(
	hashPrefix string,
	imageMap ImageMap,
	ps *models.ProcessStatus,
) ImageProcessor {
	return ImageProcessor{hashPrefix, imageMap, ps, time.Now()}
}

func (ip ImageProcessor) Process(dir string, hashLen int) error {
	start := time.Now()
	mapLen := len(ip.imageMap)
	ip.processStatus.MaxHashProgress = int32(mapLen)
	ip.processStatus.TotalImages = int32(mapLen)
	if mapLen == 0 {
		return ErrNoImages
	}

	queueSize := mapLen
	if queueSize < 10 {
		queueSize = 10
	}

	hi := HashResult{}

	hasher, err := NewHasher(HasherConfig{
		Length:     hashLen,
		Threads:    runtime.NumCPU(),
		QueueSize:  queueSize,
		HashResult: &hi,
		Prefix:     ip.hashPrefix,
	})
	if err != nil {
		return err
	}

	for fileName, cacheStatus := range ip.imageMap {
		hasher.Hash(fileName, cacheStatus, fPath.Join(dir, fileName), func(cs CacheStatus) {
			if cs == Cached {
				ip.processStatus.IncCachedImages()
			}
			ip.processStatus.IncHashProgress()
		})
	}

	hasher.Wait()
	ip.processStatus.HashingTook = time.Since(start)
	start = time.Now()
	fi := &FilteredImages{}
	imgFilter := NewImageFilter()
	imgFilter.Filter(hi, fi)
	ip.processStatus.FilterTook = time.Since(start)
	return ip.updateImages(*fi)
}

func (ip ImageProcessor) updateImages(fi FilteredImages) error {
	start := time.Now()

	dupeLen := len(fi.dupeImagePaths)
	renameLen := len(fi.imagePathMap)
	if dupeLen == 0 && renameLen == 0 {
		ip.processStatus.TotalTime = time.Since(ip.processStartTime)
		return nil
	}

	ip.processStatus.DupeImages = int32(dupeLen)
	ip.processStatus.NewImages = int32(renameLen)
	workLen := len(fi.dupeImagePaths) + len(fi.imagePathMap)
	ip.processStatus.MaxUpdateProgress = int32(workLen)
	queueSize := workLen
	if queueSize < 10 {
		queueSize = 10
	}

	tp, err := utils.NewThreadPool(runtime.NumCPU(), queueSize, false)
	if err != nil {
		return err
	}

	errors := []error{}
	mux := sync.Mutex{}

	for _, imgPath := range fi.dupeImagePaths {
		tp.Queue(func() {
			err := os.Remove(imgPath)
			if err != nil {
				mux.Lock()
				errors = append(errors, err)
				mux.Unlock()
			}
			ip.processStatus.IncUpdateProgress()
		})
	}

	for newImgHash, imgPath := range fi.imagePathMap {
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
			ip.processStatus.IncUpdateProgress()
		})
	}

	tp.Wait()
	ip.processStatus.UpdatingTook = time.Since(start)
	ip.processStatus.TotalTime = time.Since(ip.processStartTime)
	if len(errors) > 0 {
		return fmt.Errorf("update errors: %v", errors)
	}
	return nil
}
