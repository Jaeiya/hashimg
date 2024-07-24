package lib

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/jaeiya/go-template/internal"
)

const hashPrefix = "0x@"

var validImgExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
}

type FileHashInfo struct {
	hash string
	// Path to the file
	path string
	// If image has already been hashed and prefixed
	cached bool
	// If an error occurs during processing
	err error
}

func ProcessImages(dir string) error {
	fileNames, err := getImgFileNames(dir)
	if err != nil {
		return err
	}

	if len(fileNames) == 0 {
		return fmt.Errorf("no images found in %s", dir)
	}

	// We definitely don't need more than 100 for 5 threads, but
	// if we have less than 100 images, we don't even need 100.
	queueSize := 100
	if len(fileNames) < 100 {
		queueSize = len(fileNames)
	}

	tp := NewThreadPool(5, queueSize, make(chan FileHashInfo))

	filteredImages := FilteredImages{}
	imgFilter := NewImageFilter()
	imgFilter.FilterImages(tp.ResultChan, &filteredImages)

	for _, fn := range fileNames {
		if strings.HasPrefix(fn, hashPrefix) {
			ext := path.Ext(fn)
			h := strings.TrimPrefix(fn[0:len(fn)-len(ext)], hashPrefix)
			tp.Queue(func() FileHashInfo {
				return FileHashInfo{
					hash:   h,
					path:   path.Join(dir, fn),
					cached: true,
				}
			})
			continue
		}
		tp.Queue(func() FileHashInfo {
			file, err := os.Open(path.Join(dir, fn))
			if err != nil {
				return FileHashInfo{
					err: err,
				}
			}
			defer file.Close()
			return hash(file, path.Join(dir, fn), 24)
		})
	}

	tp.Wait()
	imgFilter.Wait()
	return updateImages(filteredImages)
}

func getImgFileNames(dir string) ([]string, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fileNames := []string{}
	for _, entry := range dirEntries {
		if entry.IsDir() || !validImgExtensions[path.Ext(entry.Name())] {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}

	return fileNames, nil
}

func hash(reader io.ReadCloser, path string, length int) FileHashInfo {
	h := sha256.New()

	if _, err := io.Copy(h, reader); err != nil {
		return FileHashInfo{
			err: err,
		}
	}

	return FileHashInfo{
		hash:   fmt.Sprintf("%x", h.Sum(nil))[0:length],
		path:   path,
		cached: false,
		err:    nil,
	}
}

func updateImages(fr FilteredImages) error {
	if len(fr.dupeImageHashes) == 0 && len(fr.newImageHashes) == 0 {
		return nil
	}

	workLen := len(fr.dupeImageHashes) + len(fr.newImageHashes)
	queueSize := 100
	if workLen < 100 {
		queueSize = workLen
	}

	tp := NewThreadPool(5, queueSize, make(chan error))

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
			ext := path.Ext(imgPath)
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
		return fmt.Errorf("update errors: %v", fErrors)
	}

	return nil
}
