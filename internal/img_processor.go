package lib

import (
	"fmt"
	"os"
	fPath "path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jaeiya/go-template/internal/models"
	"github.com/jaeiya/go-template/internal/utils"
)

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

func (ip ImageProcessor) Process(dir string, hashLen int, useAvgBufferSize bool) error {
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

	analyzeStart := time.Now()
	var bufferSize int64
	if useAvgBufferSize {
		avgBytes, err := getAvgFileSize(dir)
		if err != nil {
			return err
		}
		bufferSize = avgBytes
		ip.processStatus.AnalyzeTook = time.Since(analyzeStart)
	}

	start := time.Now()
	hr := HashResult{}
	hasher, err := NewHasher(HasherConfig{
		Length:     hashLen,
		Threads:    runtime.NumCPU(),
		QueueSize:  queueSize,
		HashResult: &hr,
		Prefix:     ip.hashPrefix,
		BufferSize: bufferSize,
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
	hashErr := verifyHashResult(hr)
	if hashErr != nil {
		ip.processStatus.HashErr = hashErr
		return hashErr
	}
	start = time.Now()
	fi := &FilteredImages{}
	imgFilter := NewImageFilter()
	imgFilter.Filter(hr, fi)
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
				ip.processStatus.UpdateErr = err
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
				ip.processStatus.UpdateErr = err
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

func verifyHashResult(hr HashResult) error {
	for _, r := range hr.newHashes {
		if r.err != nil {
			return r.err
		}
	}

	for _, r := range hr.oldHashes {
		if r.err != nil {
			return r.err
		}
	}

	return nil
}

func getAvgFileSize(dir string) (int64, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	var totalSize int64 = 0
	for _, entry := range files {
		info, err := entry.Info()
		if err != nil {
			return 0, err
		}
		totalSize += info.Size()
	}

	return totalSize / int64(len(files)), nil
}
