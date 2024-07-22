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
	dirEntries, err := dirReader.ReadDir(dir)
	if err != nil {
		return err
	}

	if len(dirEntries) == 0 {
		return fmt.Errorf("no files found in %s", dir)
	}

	// We definitely don't need more than 100 for 5 threads, but
	// if we have less than 100 images, we don't even need 100.
	queueSize := 100
	if len(dirEntries) < 100 {
		queueSize = len(dirEntries)
	}

	fResult := make(chan FilterResult, 1)
	hashChan := make(chan HashResult)
	tp := NewThreadPool(5, queueSize)

	go filterImages(hashChan, fResult)

	for _, entry := range dirEntries {
		ext := path.Ext(entry)
		if validExtensions[ext] {
			if strings.HasPrefix(entry, hashPrefix) {
				h := strings.TrimPrefix(strings.Split(entry, ".")[0], hashPrefix)
				hashChan <- HashResult{
					hash:   h,
					path:   path.Join(dir, entry),
					cached: true,
				}
				continue
			}
			tp.Queue(func() {
				file, err := fileOpener.Open(path.Join(dir, entry))
				if err != nil {
					hashChan <- HashResult{
						err: err,
					}
					return
				}
				defer file.Close()
				hasher.Hash(file, path.Join(dir, entry), hashChan)
			})
		}
	}

	tp.Wait()
	close(hashChan)
	filteredImages := <-fResult
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

	errChan := make(chan error)
	doneChan := make(chan bool)
	var errors []error

	go func() {
		for err := range errChan {
			errors = append(errors, err)
		}
		close(doneChan)
	}()

	tp := NewThreadPool(5, queueSize)
	for _, imgPath := range fr.dupeImageHashes {
		tp.Queue(func() {
			err := os.Remove(imgPath)
			if err != nil {
				errChan <- err
			}
		})
	}

	for newImgHash, imgPath := range fr.newImageHashes {
		tp.Queue(func() {
			dir := path.Dir(imgPath)
			ext := path.Ext(imgPath)
			newFileName := path.Join(dir, hashPrefix+newImgHash+ext)
			err := os.Rename(imgPath, newFileName)
			if err != nil {
				errChan <- err
			}
		})
	}

	tp.Wait()
	close(errChan)
	<-doneChan

	if len(errors) > 0 {
		return fmt.Errorf("update errors: %v", errors)
	}

	return nil
}
