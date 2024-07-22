package lib

import "os"

type ImgDirReader interface {
	ReadDir(dir string) ([]string, error)
}

type MyDirReader struct{}

/*
ReadDir returns the list of files in the specified directory
*/
func (dr MyDirReader) ReadDir(dir string) ([]string, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var entries []string
	for _, entry := range dirEntries {
		entries = append(entries, entry.Name())
	}

	return entries, nil
}
