package lib

import (
	"fmt"
	"os"
	fp "path/filepath"
	"strings"
	"sync"
)

var validImgExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
	".heic": true,
}

type ProcessStats struct {
	Total int
	New   int
	Dup   int
}

func ProcessImages(dir string) (ProcessStats, error) {
	fileNames, err := getImgFileNames(dir)
	if err != nil {
		return ProcessStats{}, err
	}

	if len(fileNames) == 0 {
		return ProcessStats{}, fmt.Errorf("no images found in %s", dir)
	}

	queueSize := len(fileNames)
	if queueSize < 10 {
		queueSize = 10
	}

	hi := []HashInfo{}
	hasher := NewHasher(24, 10, &hi)

	for _, fn := range fileNames {
		hasher.Hash(fn, fp.Join(dir, fn))
	}

	hasher.Wait()
	fi := &FilteredImages{}
	imgFilter := NewImageFilter()
	imgFilter.FilterImages(hi, fi)
	return updateImages(*fi)
}

func getImgFileNames(dir string) ([]string, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fileNames := []string{}
	for _, entry := range dirEntries {
		if entry.IsDir() || !validImgExtensions[strings.ToLower(fp.Ext(entry.Name()))] {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}

	return fileNames, nil
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
			dir := fp.Dir(imgPath)
			// Extensions should always be lowercase even though the
			// file system doesn't care
			ext := strings.ToLower(fp.Ext(imgPath))
			newFileName := fp.Join(dir, hashPrefix+newImgHash+ext)
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
