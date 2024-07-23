package lib

import (
	"fmt"
	"os"
	"path"
	"strings"
)

const hashPrefix = "0x@"

var validExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
}

func ProcessImages(dir string) error {
	return process(MyDirReader{}, MyHasher{}, MyFileOpener{}, dir)
}

func process(
	dirReader ImgDirReader,
	hasher ImgHasher,
	fileOpener ImgFileOpener,
	dir string,
) error {
	fileNames, err := dirReader.ReadDir(dir)
	if err != nil {
		return err
	}

	if len(fileNames) == 0 {
		return fmt.Errorf("no files found in %s", dir)
	}

	// We definitely don't need more than 100 for 5 threads, but
	// if we have less than 100 images, we don't even need 100.
	queueSize := 100
	if len(fileNames) < 100 {
		queueSize = len(fileNames)
	}

	fResultChan := make(chan FilterResult, 1)
	tp := NewThreadPool(5, queueSize, make(chan HashResult))

	go filterImages(tp.ResultChan, fResultChan)

	for _, fn := range fileNames {
		ext := path.Ext(fn)
		if validExtensions[ext] {
			if strings.HasPrefix(fn, hashPrefix) {
				h := strings.TrimPrefix(strings.Split(fn, ".")[0], hashPrefix)
				tp.Queue(func() HashResult {
					return HashResult{
						hash:   h,
						path:   path.Join(dir, fn),
						cached: true,
					}
				})
				continue
			}
			tp.Queue(func() HashResult {
				file, err := fileOpener.Open(path.Join(dir, fn))
				if err != nil {
					return HashResult{
						err: err,
					}
				}
				defer file.Close()
				return hasher.Hash(file, path.Join(dir, fn), 24)
			})
		}
	}

	tp.Wait()
	filteredImages := <-fResultChan
	return UpdateImages(filteredImages)
}

func filterImages(hResult chan HashResult, fResult chan<- FilterResult) {
	oldImageHashes := map[string]string{}
	newImageHashes := map[string]string{}
	dupeImageHashes := []string{}

	for hr := range hResult {
		if hr.err != nil {
			fmt.Println(hr.err)
			continue
		}

		if hr.cached {
			oldImageHashes[hr.hash] = hr.path
			continue
		}

		if _, ok := newImageHashes[hr.hash]; ok {
			dupeImageHashes = append(dupeImageHashes, hr.path)
			continue
		}

		newImageHashes[hr.hash] = hr.path
	}

	for oldImgHash := range oldImageHashes {
		if imgPath, ok := newImageHashes[oldImgHash]; ok {
			dupeImageHashes = append(dupeImageHashes, imgPath)
			delete(newImageHashes, oldImgHash)
			continue
		}
	}

	fResult <- FilterResult{
		newImageHashes:  newImageHashes,
		dupeImageHashes: dupeImageHashes,
	}
}

func UpdateImages(fr FilterResult) error {
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

	fErrors := filterNils(errors)

	if len(fErrors) > 0 {
		return fmt.Errorf("update errors: %v", fErrors)
	}

	return nil
}

func filterNils(errs []error) []error {
	var r []error
	for _, v := range errs {
		if v != nil {
			r = append(r, v)
		}
	}
	return r
}
