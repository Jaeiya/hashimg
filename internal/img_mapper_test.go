package internal

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockMapperTest struct {
	should      string
	files       []string
	fileContent []string
	expectMap   ImageMap
}

func TestImageMapper(t *testing.T) {
	const hashPrefix = "0x@"
	t.Run("should error if directory not found", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		_, err := MapImages("", hashPrefix)
		a.ErrorContains(err, "system cannot find the file")
	})

	t.Run("should error if directory is empty", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		dir := t.TempDir()
		_, err := MapImages(filepath.Join(dir), hashPrefix)
		a.ErrorIs(err, ErrNoImages)
	})

	t.Run("should error if no image files found", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		dir := t.TempDir()
		err := writeFiles(dir, []string{"test1.txt", "test2.mp3"}, []string{"test1", "test2"})
		a.NoError(err)
		_, err = MapImages(filepath.Join(dir), hashPrefix)
		a.ErrorIs(err, ErrNoImages)
	})

	mockTable := []MockMapperTest{
		{
			should: "map cached images",
			files: []string{
				"0x@1b4f0e9851971998e7320785.png",
				"0x@60303ae22b998861bce3b28f.png",
			},
			fileContent: []string{"test1", "test2"},
			expectMap: ImageMap{
				"0x@1b4f0e9851971998e7320785.png": Cached,
				"0x@60303ae22b998861bce3b28f.png": Cached,
			},
		},
		{
			should: "map non cached images",
			files: []string{
				"test1.png",
				"test2.jpg",
			},
			fileContent: []string{"test1", "test2"},
			expectMap: ImageMap{
				"test1.png": NotCached,
				"test2.jpg": NotCached,
			},
		},
		{
			should: "detect both cached and non-cached images",
			files: []string{
				"0x@1b4f0e9851971998e7320785.png",
				"test2.bmp",
				"0x@fd61a03af4f77d870fc21e05.gif",
				"test4.jpg",
			},
			fileContent: []string{"test1", "test2", "test3", "test4"},
			expectMap: ImageMap{
				"0x@1b4f0e9851971998e7320785.png": Cached,
				"test2.bmp":                       NotCached,
				"0x@fd61a03af4f77d870fc21e05.gif": Cached,
				"test4.jpg":                       NotCached,
			},
		},
	}

	for _, test := range mockTable {
		t.Run(test.should, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			dir := t.TempDir()
			err := writeFiles(dir, test.files, test.fileContent)
			a.NoError(err)
			iMap, err := MapImages(dir, hashPrefix)
			a.NoError(err)
			a.Equal(test.expectMap, iMap, "expected mapped results should match actual")
		})
	}
}
