package internal

import (
	"fmt"
)

type FilteredImages struct {
	imagePathMap   map[string]string
	dupeImagePaths []string
}

func NewImageFilter() ImageFilter {
	return ImageFilter{}
}

type ImageFilter struct{}

func (ImageFilter) Filter(hashResult HashResult, fi *FilteredImages) {
	imagePathMap := map[string]string{}
	dupeImagePaths := []string{}

	for _, hashInfo := range hashResult.newHashes {
		if hashInfo.err != nil {
			fmt.Println(hashInfo.err)
			continue
		}

		_, isOldDupe := hashResult.oldHashes[hashInfo.hash]
		_, isNewDupe := imagePathMap[hashInfo.hash]

		if isOldDupe || isNewDupe {
			dupeImagePaths = append(dupeImagePaths, hashInfo.path)
			continue
		}

		imagePathMap[hashInfo.hash] = hashInfo.path
	}

	fi.imagePathMap = imagePathMap
	fi.dupeImagePaths = dupeImagePaths
}
