package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jaeiya/hashimg/internal/models"
	"github.com/jaeiya/hashimg/internal/utils"
)

type ImageProcessor struct {
	Status           *models.ProcessStatus
	WorkingDir       string
	HasDupes         bool
	isReviewProcess  bool
	dupeReviewFolder string
	DupeRestorePaths []string
	hashPrefix       string
	hashLength       int
	imageMap         ImageMap
	ProcessTime      time.Duration
	FilteredImages   *FilteredImages
}

type ImageProcessorConfig struct {
	Prefix           string
	HashLength       int
	WorkingDir       string
	ImageMap         ImageMap
	DupeReviewFolder string
}

type FilteredImages struct {
	NewHashesMap map[string]HashInfo
	DupeMap      map[string][]HashInfo
}

func NewImageProcessor(cfg ImageProcessorConfig) *ImageProcessor {
	if cfg.DupeReviewFolder == "" {
		cfg.DupeReviewFolder = "__dupes"
	}
	return &ImageProcessor{
		Status:           &models.ProcessStatus{},
		WorkingDir:       cfg.WorkingDir,
		dupeReviewFolder: filepath.Join(cfg.WorkingDir, cfg.DupeReviewFolder),
		hashPrefix:       cfg.Prefix,
		hashLength:       cfg.HashLength,
		imageMap:         cfg.ImageMap,
		DupeRestorePaths: []string{},
	}
}

func (ip *ImageProcessor) ProcessAll(useBuffer bool) error {
	err := ip.ProcessHash(useBuffer)
	if err != nil {
		return err
	}
	return ip.Update()
}

/*
ProcessHashReview does the same thing as ProcessHash() but also moves
all duplicate images to a temporary folder, for the user to review.
The folder is automatically opened, after the images are moved.
*/
func (ip *ImageProcessor) ProcessHashReview(useBuffer bool) error {
	err := ip.ProcessHash(useBuffer)
	if err != nil {
		return err
	}

	if !ip.HasDupes {
		return nil
	}

	ip.isReviewProcess = true
	fi := ip.FilteredImages

	err = os.MkdirAll(ip.dupeReviewFolder, os.ModeDir|os.ModeAppend)
	if err != nil {
		return err
	}

	cachedImageCount := 0
	for _, dupes := range fi.DupeMap {
		for i, dupe := range dupes {
			ext := filepath.Ext(dupe.path)
			reviewFileName := fmt.Sprintf("%s_%d%s", dupe.hash, i+1, ext)
			err = os.Rename(
				dupe.path,
				filepath.Join(ip.dupeReviewFolder, reviewFileName),
			)
			if err != nil {
				return err
			}
			// cached images are not "new"
			if dupe.cached {
				cachedImageCount += 1
			}
			if dupe.isNovel {
				ip.DupeRestorePaths = append(
					ip.DupeRestorePaths,
					filepath.Join(ip.dupeReviewFolder, reviewFileName),
				)
				dupe.path = filepath.Join(ip.WorkingDir, reviewFileName)
				fi.NewHashesMap[dupe.hash] = dupe
				continue
			}
			ip.Status.DupeImageCount += 1
		}
	}

	ip.Status.NewImageCount = int32(len(fi.NewHashesMap) - cachedImageCount)

	utils.OpenFolder(ip.dupeReviewFolder)

	return nil
}

func (ip *ImageProcessor) RestoreFromReview() error {
	for _, path := range ip.DupeRestorePaths {
		err := os.Rename(
			path,
			filepath.Join(ip.WorkingDir, filepath.Base(path)),
		)
		if err != nil {
			return err
		}
	}
	return os.RemoveAll(ip.dupeReviewFolder)
}

/*
ProcessHash calculates the hash of all images in the image map and
separates out the duplicates from the new images. The result is
saved to the filteredImages field.
*/
func (ip *ImageProcessor) ProcessHash(useBuffer bool) error {
	timeStart := time.Now()
	defer func() { ip.ProcessTime = time.Since(timeStart) }()

	if len(ip.imageMap) == 0 {
		return ErrNoImages
	}

	ip.Status.TotalImageCount = int32(len(ip.imageMap))
	ip.Status.MaxHashProgress = ip.Status.TotalImageCount

	bufferSize, err := ip.calcBufferSize(useBuffer)
	if err != nil {
		ip.Status.HashErr = err
		return err
	}

	hashResult, err := ip.hashImages(bufferSize)
	if err != nil {
		ip.Status.HashErr = err
		return err
	}

	ip.FilteredImages = ip.filterImages(hashResult)
	return nil
}

/*
Update handles renaming and deleting images that have had their hashes
processed. If the type of process was a "review", then it ONLY renames
the images.
*/
func (ip *ImageProcessor) Update() error {
	timeStart := time.Now()
	defer func() { ip.ProcessTime = ip.ProcessTime + time.Since(timeStart) }()

	if ip.isReviewProcess {
		if err := ip.renameOnly(); err != nil {
			ip.Status.UpdateErr = err
			return err
		}
		return nil
	}

	if err := ip.deleteAndRename(); err != nil {
		ip.Status.UpdateErr = err
		return err
	}

	return nil
}

func (ip *ImageProcessor) calcBufferSize(useBuffer bool) (int64, error) {
	if !useBuffer {
		return 0, nil
	}

	start := time.Now()
	defer func() { ip.Status.AnalyzeTook = time.Since(start) }()

	files, err := os.ReadDir(ip.WorkingDir)
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

func (ip *ImageProcessor) hashImages(bufferSize int64) (HashResult, error) {
	start := time.Now()
	defer func() { ip.Status.HashingTook = time.Since(start) }()

	hr := HashResult{}

	hasher, err := NewHasher(HasherConfig{
		Length:     ip.hashLength,
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
		hasher.Hash(
			fileName,
			cacheStatus,
			filepath.Join(ip.WorkingDir, fileName),
			func(cs CacheStatus) {
				if cs == Cached {
					ip.Status.IncCachedImages()
				}
				ip.Status.IncHashProgress()
			},
		)
	}

	hasher.Wait()

	if hashErr := verifyHashResult(hr); hashErr != nil {
		return hr, hashErr
	}

	return hr, nil
}

func (ip *ImageProcessor) filterImages(hr HashResult) *FilteredImages {
	start := time.Now()
	defer func() { ip.Status.FilterTook = time.Since(start) }()

	newImagesMap := map[string]HashInfo{}
	dupeImageMap := map[string][]HashInfo{}

	for _, hashInfo := range hr.newHashesInfo {
		if _, isDupe := dupeImageMap[hashInfo.hash]; isDupe {
			dupeImageMap[hashInfo.hash] = append(dupeImageMap[hashInfo.hash], hashInfo)
			delete(newImagesMap, hashInfo.hash)
			continue
		}

		if oldInfo, isDupe := hr.oldHashesInfo[hashInfo.hash]; isDupe {
			oldInfo.isNovel = true
			dupeImageMap[oldInfo.hash] = []HashInfo{oldInfo, hashInfo}
			delete(newImagesMap, hashInfo.hash)
			continue
		}

		if newInfo, isDupe := newImagesMap[hashInfo.hash]; isDupe {
			newInfo.isNovel = true
			dupeImageMap[newInfo.hash] = []HashInfo{newInfo, hashInfo}
			delete(newImagesMap, hashInfo.hash)
			continue
		}

		newImagesMap[hashInfo.hash] = hashInfo
	}

	ip.HasDupes = len(dupeImageMap) > 0

	return &FilteredImages{
		NewHashesMap: newImagesMap,
		DupeMap:      dupeImageMap,
	}
}

func (ip *ImageProcessor) deleteAndRename() error {
	if ip.isReviewProcess {
		return fmt.Errorf("you must use renameOnly() when using review process")
	}

	if ip.FilteredImages == nil {
		return fmt.Errorf("cannot update images: hashes have not been processed")
	}

	start := time.Now()
	fi := ip.FilteredImages
	defer func() {
		ip.Status.UpdatingTook = time.Since(start)
	}()

	if len(fi.DupeMap) == 0 && len(fi.NewHashesMap) == 0 {
		return nil
	}

	for _, dupes := range fi.DupeMap {
		for i, dupe := range dupes {
			if dupe.cached {
				continue
			}
			// Save all novel images
			if dupe.isNovel {
				fi.DupeMap[dupe.hash] = dupes[i+1:]
				fi.NewHashesMap[dupe.hash] = dupe
				continue
			}
			ip.Status.DupeImageCount += 1
		}
	}

	ip.Status.NewImageCount = int32(len(fi.NewHashesMap))
	ip.Status.MaxUpdateProgress = ip.Status.DupeImageCount + int32(len(fi.NewHashesMap))

	tp, err := utils.NewThreadPool(
		runtime.NumCPU(),
		max(len(fi.DupeMap)+len(fi.NewHashesMap), 10),
		false,
	)
	if err != nil {
		return err
	}

	errors := []error{}
	mux := sync.Mutex{}

	for _, dupes := range fi.DupeMap {
		for _, dupe := range dupes {
			if dupe.cached {
				continue
			}
			tp.Queue(func() {
				err := os.Remove(dupe.path)
				if err != nil {
					mux.Lock()
					errors = append(errors, err)
					mux.Unlock()
				}
				ip.Status.IncUpdateProgress()
			})
		}
	}

	for newImgHash, hashInfo := range fi.NewHashesMap {
		tp.Queue(func() {
			dir := filepath.Dir(hashInfo.path)
			// Uppercase extensions are ugly and inconsistent
			ext := strings.ToLower(filepath.Ext(hashInfo.path))
			newFileName := filepath.Join(dir, ip.hashPrefix+newImgHash+ext)
			err := os.Rename(hashInfo.path, newFileName)
			if err != nil {
				mux.Lock()
				errors = append(errors, err)
				mux.Unlock()
			}
			ip.Status.IncUpdateProgress()
		})
	}

	tp.Wait()

	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

func (ip *ImageProcessor) renameOnly() error {
	fi := ip.FilteredImages

	tp, err := utils.NewThreadPool(
		runtime.NumCPU(),
		max(len(fi.DupeMap)+len(fi.NewHashesMap), 10),
		false,
	)
	if err != nil {
		return err
	}

	mux := sync.Mutex{}
	errors := []error{}
	ip.Status.MaxUpdateProgress = int32(len(fi.NewHashesMap))

	for newImgHash, hashInfo := range fi.NewHashesMap {
		tp.Queue(func() {
			dir := filepath.Dir(hashInfo.path)
			// Uppercase extensions are ugly and inconsistent
			ext := strings.ToLower(filepath.Ext(hashInfo.path))
			newFileName := filepath.Join(dir, ip.hashPrefix+newImgHash+ext)
			err := os.Rename(hashInfo.path, newFileName)
			if err != nil {
				mux.Lock()
				errors = append(errors, err)
				mux.Unlock()
			}
			ip.Status.IncUpdateProgress()
		})
	}

	tp.Wait()

	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

func verifyHashResult(hr HashResult) error {
	for _, r := range hr.newHashesInfo {
		if r.err != nil {
			return r.err
		}
	}

	for _, r := range hr.oldHashesInfo {
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
