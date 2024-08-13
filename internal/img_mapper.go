package internal

import (
	"os"
	fPath "path/filepath"
	"strings"
)

const (
	NotCached   CacheStatus  = false
	Cached      CacheStatus  = true
	ExtEnabled  ImgExtStatus = true
	ExtDisabled ImgExtStatus = false
)

var imageExtensions = ImageExtMap{
	".jpg":  ExtEnabled,
	".jpeg": ExtEnabled,
	".png":  ExtEnabled,
	".gif":  ExtEnabled,
	".bmp":  ExtEnabled,
	".webp": ExtEnabled,
	".heic": ExtEnabled,
}

type (
	CacheStatus  bool
	ImgExtStatus bool
)

type ImageMap map[string]CacheStatus

type ImageExtMap map[string]ImgExtStatus

func MapImages(dir, hashPrefix string) (ImageMap, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	if len(dirEntries) == 0 {
		return nil, ErrNoImages
	}

	iMap := ImageMap{}
	for _, entry := range dirEntries {
		fileName := entry.Name()
		// Some extensions might be uppercase
		imgExt := strings.ToLower(fPath.Ext(fileName))
		if entry.IsDir() || imageExtensions[imgExt] == ExtDisabled {
			continue
		}

		if strings.HasPrefix(fileName, hashPrefix) {
			iMap[fileName] = Cached
		} else {
			iMap[fileName] = NotCached
		}
	}

	if len(iMap) == 0 {
		return nil, ErrNoImages
	}

	return iMap, nil
}
