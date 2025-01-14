package lib

import (
	"os"
	fPath "path/filepath"
	"strings"
)

type (
	CacheStatus bool
	ImageMap    map[string]CacheStatus
	ExtState    bool
	ImageExtMap map[string]ExtState
)

const (
	NotCached   CacheStatus = false
	Cached      CacheStatus = true
	ExtEnabled  ExtState    = true
	ExtDisabled ExtState    = false
)

var imageExtensions = ImageExtMap{
	".apng": ExtEnabled,
	".avif": ExtEnabled,
	".bmp":  ExtEnabled,
	".gif":  ExtEnabled,
	".heic": ExtEnabled,
	".heif": ExtEnabled,
	".jpg":  ExtEnabled,
	".jpeg": ExtEnabled,
	".png":  ExtEnabled,
	".svg":  ExtEnabled,
	".tif":  ExtEnabled,
	".tiff": ExtEnabled,
	".webp": ExtEnabled,
}

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
