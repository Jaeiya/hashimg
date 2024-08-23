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

var mux = sync.Mutex{}

type ImageProcessor struct {
	Status           *models.ProcessStatus
	WorkingDir       string
	HasDupes         bool
	OpenReviewFolder bool
	isReviewProcess  bool
	dupeReviewFolder string
	DupeRestorePaths []string
	hashPrefix       string
	hashLength       int
	imageMap         ImageMap
	ProcessTime      time.Duration
	processedImages  *ProcessedImages
}

type ImageProcessorConfig struct {
	Prefix           string
	HashLength       int
	WorkingDir       string
	ImageMap         ImageMap
	DupeReviewFolder string
	OpenReviewFolder bool
}

type ProcessedImages struct {
	NewImagesByHash  map[string]HashInfo
	DupeImagesByHash map[string][]HashInfo
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
		OpenReviewFolder: cfg.OpenReviewFolder,
	}
}

func (ip *ImageProcessor) ProcessAll(useBuffer bool) error {
	err := ip.ProcessImages(useBuffer)
	if err != nil {
		return err
	}
	return ip.UpdateImages()
}

/*
ProcessImagesForReview does the same thing as ProcessImages(), but also moves
all duplicate images to a temporary folder for the user to review. The folder
is automatically opened after the images are moved.
*/
func (ip *ImageProcessor) ProcessImagesForReview(useBuffer bool) error {
	err := ip.ProcessImages(useBuffer)
	if err != nil {
		return err
	}

	if !ip.HasDupes {
		return nil
	}

	ip.isReviewProcess = true
	pi := ip.processedImages

	err = os.MkdirAll(ip.dupeReviewFolder, os.ModeDir|os.ModeAppend)
	if err != nil {
		return err
	}

	cachedImageCount := 0
	for _, dupes := range pi.DupeImagesByHash {
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
				pi.NewImagesByHash[dupe.hash] = dupe
				continue
			}
			ip.Status.DupeImageCount += 1
		}
	}

	ip.Status.NewImageCount = int32(len(pi.NewImagesByHash) - cachedImageCount)

	if ip.OpenReviewFolder {
		utils.OpenFolder(ip.dupeReviewFolder)
	}

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
ProcessImages calculates the hash of all images in the image map and
separates out the duplicates from the new images.

🟡 The result is saved to a field within the image processor.
This allows us to call UpdateImages() without a dependency.
*/
func (ip *ImageProcessor) ProcessImages(useBuffer bool) error {
	timeStart := time.Now()
	defer func() {
		ip.ProcessTime = time.Since(timeStart)
		ip.Status.ProcessingComplete = true
	}()

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

	hashResult, err := ip.calcImageHashes(bufferSize)
	if err != nil {
		ip.Status.HashErr = err
		return err
	}

	start := time.Now()
	defer func() { ip.Status.FilterTook = time.Since(start) }()

	newImagesByHash := map[string]HashInfo{}
	dupeImagesByHash := map[string][]HashInfo{}

	for _, hashInfo := range hashResult.newHashesInfo {
		if _, isDupe := dupeImagesByHash[hashInfo.hash]; isDupe {
			dupeImagesByHash[hashInfo.hash] = append(dupeImagesByHash[hashInfo.hash], hashInfo)
			delete(newImagesByHash, hashInfo.hash)
			continue
		}

		if oldInfo, isDupe := hashResult.oldHashesInfo[hashInfo.hash]; isDupe {
			oldInfo.isNovel = true
			dupeImagesByHash[oldInfo.hash] = []HashInfo{oldInfo, hashInfo}
			delete(newImagesByHash, hashInfo.hash)
			continue
		}

		if newInfo, isDupe := newImagesByHash[hashInfo.hash]; isDupe {
			newInfo.isNovel = true
			dupeImagesByHash[newInfo.hash] = []HashInfo{newInfo, hashInfo}
			delete(newImagesByHash, hashInfo.hash)
			continue
		}

		newImagesByHash[hashInfo.hash] = hashInfo
	}

	ip.HasDupes = len(dupeImagesByHash) > 0

	ip.processedImages = &ProcessedImages{newImagesByHash, dupeImagesByHash}
	return nil
}

/*
UpdateImages handles renaming and deleting images that have had their hashes
processed. If the type of process was a "review", then it ONLY renames
the images.
*/
func (ip *ImageProcessor) UpdateImages() error {
	timeStart := time.Now()
	defer func() {
		ip.ProcessTime = ip.ProcessTime + time.Since(timeStart)
		ip.Status.UpdatingComplete = true
	}()

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

func (ip *ImageProcessor) calcImageHashes(bufferSize int64) (HashResult, error) {
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

func (ip *ImageProcessor) deleteAndRename() error {
	if ip.isReviewProcess {
		return fmt.Errorf("you must use renameOnly() when using review process")
	}

	if ip.processedImages == nil {
		return fmt.Errorf("cannot update images: hashes have not been processed")
	}

	start := time.Now()
	defer func() {
		ip.Status.UpdatingTook = time.Since(start)
	}()

	newImages := ip.processedImages.NewImagesByHash
	dupeImages := ip.processedImages.DupeImagesByHash

	if len(dupeImages) == 0 && len(newImages) == 0 {
		return nil
	}

	for _, dupes := range dupeImages {
		for i, dupe := range dupes {
			if dupe.cached {
				continue
			}
			if dupe.isNovel {
				// All novel images are at index 0
				dupeImages[dupe.hash] = dupes[i+1:]
				newImages[dupe.hash] = dupe
				continue
			}
			ip.Status.DupeImageCount += 1
		}
	}

	ip.Status.NewImageCount = int32(len(newImages))
	ip.Status.MaxUpdateProgress = ip.Status.DupeImageCount + int32(len(newImages))

	tp, err := utils.NewThreadPool(
		runtime.NumCPU(),
		max(len(dupeImages)+len(newImages), 10),
		false,
	)
	if err != nil {
		return err
	}

	errors := []error{}

	for _, dupes := range dupeImages {
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

	for newImgHash, hashInfo := range newImages {
		tp.Queue(func() {
			err := ip.renameImages(hashInfo, newImgHash)
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
	pi := ip.processedImages

	tp, err := utils.NewThreadPool(
		runtime.NumCPU(),
		max(len(pi.DupeImagesByHash)+len(pi.NewImagesByHash), 10),
		false,
	)
	if err != nil {
		return err
	}

	errors := []error{}
	ip.Status.MaxUpdateProgress = int32(len(pi.NewImagesByHash))

	for newImgHash, hashInfo := range pi.NewImagesByHash {
		tp.Queue(func() {
			err := ip.renameImages(hashInfo, newImgHash)
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

func (ip *ImageProcessor) renameImages(hi HashInfo, newImgHash string) error {
	dir := filepath.Dir(hi.path)
	// Uppercase extensions are ugly and inconsistent
	ext := strings.ToLower(filepath.Ext(hi.path))
	newFileName := filepath.Join(dir, ip.hashPrefix+newImgHash+ext)
	err := os.Rename(hi.path, newFileName)
	if err != nil {
		return err
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
