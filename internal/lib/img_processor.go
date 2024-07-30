package lib

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/jaeiya/go-template/internal"
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
		hasher.Hash(fn, path.Join(dir, fn))
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
		if entry.IsDir() || !validImgExtensions[strings.ToLower(path.Ext(entry.Name()))] {
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
	queueSize := 100
	if workLen < 100 {
		if workLen < 10 {
			queueSize = 10
		} else {
			queueSize = workLen
		}
	}

	tp := NewThreadPool[error](10, queueSize, true)

	var errors []error
	go func() {
		for err := range tp.ResultChan {
			errors = append(errors, err)
		}
	}()

	for _, imgPath := range fr.dupeImageHashes {
		tp.Queue(func() error {
			err := os.Remove(imgPath)
			if err != nil {
				return err
			}
			return nil
		})
	}

	for newImgHash, imgPath := range fr.newImageHashes {
		tp.Queue(func() error {
			dir := path.Dir(imgPath)
			// Extensions should always be lowercase even though the
			// file system doesn't care
			ext := strings.ToLower(path.Ext(imgPath))
			newFileName := path.Join(dir, hashPrefix+newImgHash+ext)
			err := os.Rename(imgPath, newFileName)
			if err != nil {
				return err
			}
			return nil
		})
	}

	tp.Wait()

	fErrors := internal.FilterNils(errors)
	if len(fErrors) > 0 {
		return ProcessStats{}, fmt.Errorf("update errors: %v", fErrors)
	}

	return ProcessStats{
		Total: workLen,
		New:   len(fr.newImageHashes),
		Dup:   len(fr.dupeImageHashes),
	}, nil
}
