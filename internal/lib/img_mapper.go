package lib

import (
	"os"
	fPath "path/filepath"
	"strings"
)

var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
	".heic": true,
}

type CacheStatus bool

type ImageMap map[string]CacheStatus

const (
	NotCached CacheStatus = false
	Cached    CacheStatus = true
)

func GetImageData(dir string) (ImageMap, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	iMap := ImageMap{}
	for _, entry := range dirEntries {
		if entry.IsDir() || !imageExtensions[strings.ToLower(fPath.Ext(entry.Name()))] {
			continue
		}

		if strings.HasPrefix(entry.Name(), hashPrefix) {
			iMap[entry.Name()] = Cached
		} else {
			iMap[entry.Name()] = NotCached
		}
	}

	return iMap, nil
}
