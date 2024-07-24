package lib

import (
	"fmt"
	"sync"
)

type FilteredImages struct {
	newImageHashes  map[string]string
	dupeImageHashes []string
}

func NewImageFilter() *ImageFilter {
	return &ImageFilter{}
}

type ImageFilter struct {
	wg sync.WaitGroup
}

func (imgf *ImageFilter) FilterImages(fhiChan chan FileHashInfo, fr *FilteredImages) {
	imgf.wg.Add(1)
	defer imgf.wg.Done()

	go func() {
		oldImageHashes := map[string]string{}
		newImageHashes := map[string]string{}
		dupeImageHashes := []string{}

		for fhi := range fhiChan {
			if fhi.err != nil {
				fmt.Println(fhi.err)
				continue
			}

			if fhi.cached {
				oldImageHashes[fhi.hash] = fhi.path
				continue
			}

			if _, ok := newImageHashes[fhi.hash]; ok {
				dupeImageHashes = append(dupeImageHashes, fhi.path)
				continue
			}

			newImageHashes[fhi.hash] = fhi.path
		}

		for oldImgHash := range oldImageHashes {
			if imgPath, ok := newImageHashes[oldImgHash]; ok {
				dupeImageHashes = append(dupeImageHashes, imgPath)
				delete(newImageHashes, oldImgHash)
				continue
			}
		}

		fr.newImageHashes = newImageHashes
		fr.dupeImageHashes = dupeImageHashes
	}()
}

func (imgf *ImageFilter) Wait() {
	imgf.wg.Wait()
}
