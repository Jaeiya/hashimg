package internal

import (
	"os"
	fPath "path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jaeiya/hashimg/internal/models"
	"github.com/jaeiya/hashimg/internal/utils"
)

type ImageProcessor struct {
	hashPrefix       string
	imageMap         ImageMap
	processStatus    *models.ProcessStatus
	processStartTime time.Time
}

type FilteredImages struct {
	imagePathMap   map[string]string
	dupeImagePaths []string
}

func NewImageProcessor(
	hashPrefix string,
	imageMap ImageMap,
	ps *models.ProcessStatus,
) ImageProcessor {
	return ImageProcessor{hashPrefix, imageMap, ps, time.Now()}
}

func (ip ImageProcessor) Process(dir string, hashLen int, useAvgBufferSize bool) error {
	if len(ip.imageMap) == 0 {
		return ErrNoImages
	}

	ip.processStatus.TotalImages = int32(len(ip.imageMap))
	ip.processStatus.MaxHashProgress = ip.processStatus.TotalImages

	bufferSize, err := ip.calcBufferSize(dir, useAvgBufferSize)
	if err != nil {
		ip.processStatus.HashErr = err
		return err
	}

	hashResult, err := ip.hashImages(dir, hashLen, bufferSize)
	if err != nil {
		ip.processStatus.HashErr = err
		return err
	}

	fi := ip.filterImages(hashResult)

	if err := ip.updateImages(fi); err != nil {
		ip.processStatus.UpdateErr = err
		return err
	}

	return nil
}

func (ip ImageProcessor) calcBufferSize(dir string, useAvgBufferSize bool) (int64, error) {
	if !useAvgBufferSize {
		return 0, nil
	}

	start := time.Now()
	defer func() { ip.processStatus.AnalyzeTook = time.Since(start) }()

	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	var totalSize int64
	for _, entry := range files {
		info, err := entry.Info()
		if err != nil {
			return 0, err
		}
		totalSize += info.Size()
	}

	return totalSize / int64(len(files)), nil
}

func (ip ImageProcessor) hashImages(
	dir string,
	hashLen int,
	bufferSize int64,
) (HashResult, error) {
	start := time.Now()
	defer func() { ip.processStatus.HashingTook = time.Since(start) }()

	hr := HashResult{}

	hasher, err := NewHasher(HasherConfig{
		Length:     hashLen,
		Threads:    runtime.NumCPU(),
		QueueSize:  max(len(ip.imageMap), 10),
		HashResult: &hr,
		Prefix:     ip.hashPrefix,
		BufferSize: bufferSize,
	})
	if err != nil {
		return hr, err
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

	if hashErr := verifyHashResult(hr); hashErr != nil {
		return hr, hashErr
	}

	return hr, nil
}

func (ip ImageProcessor) filterImages(hr HashResult) FilteredImages {
	start := time.Now()
	defer func() { ip.processStatus.FilterTook = time.Since(start) }()

	imagePathMap := map[string]string{}
	dupeImagePaths := []string{}

	for _, hashInfo := range hr.newHashes {
		_, isOldDupe := hr.oldHashes[hashInfo.hash]
		_, isNewDupe := imagePathMap[hashInfo.hash]

		if isOldDupe || isNewDupe {
			dupeImagePaths = append(dupeImagePaths, hashInfo.path)
			continue
		}

		imagePathMap[hashInfo.hash] = hashInfo.path
	}

	return FilteredImages{
		imagePathMap:   imagePathMap,
		dupeImagePaths: dupeImagePaths,
	}
}

func (ip ImageProcessor) updateImages(fi FilteredImages) error {
	start := time.Now()
	defer func() {
		ip.processStatus.UpdatingTook = time.Since(start)
		ip.processStatus.TotalTime = time.Since(ip.processStartTime)
	}()

	if len(fi.dupeImagePaths) == 0 && len(fi.imagePathMap) == 0 {
		return nil
	}

	ip.processStatus.DupeImages = int32(len(fi.dupeImagePaths))
	ip.processStatus.NewImages = int32(len(fi.imagePathMap))
	ip.processStatus.MaxUpdateProgress = int32(len(fi.dupeImagePaths) + len(fi.imagePathMap))

	tp, err := utils.NewThreadPool(
		runtime.NumCPU(),
		max(len(fi.dupeImagePaths)+len(fi.imagePathMap), 10),
		false,
	)
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

	if len(errors) > 0 {
		return errors[0]
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
