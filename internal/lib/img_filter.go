package lib

import (
	"fmt"
)

type FilteredImages struct {
	newImageHashes  map[string]string
	dupeImageHashes []string
}

func NewImageFilter() ImageFilter {
	return ImageFilter{}
}

type ImageFilter struct{}

func (ImageFilter) FilterImages(hashInfo []HashInfo, fr *FilteredImages) {
	oldImageHashes := map[string]string{}
	newImageHashes := map[string]string{}
	dupeImageHashes := []string{}

	for _, h := range hashInfo {
		if h.err != nil {
			fmt.Println(h.err)
			continue
		}

		if h.cached {
			oldImageHashes[h.hash] = h.path
			continue
		}

		if _, ok := newImageHashes[h.hash]; ok {
			dupeImageHashes = append(dupeImageHashes, h.path)
			continue
		}

		newImageHashes[h.hash] = h.path
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
}
