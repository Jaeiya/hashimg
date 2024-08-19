package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockImgProcData struct {
	should      string
	files       []string
	fileContent []string
	expectFiles []string
}

func TestImageProcessor(t *testing.T) {
	const hashPrefix = "0x@"

	t.Run("should error with an empty image map", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		wd, _ := os.Getwd()
		imgProcessor := NewImageProcessor(ImageProcessorConfig{
			WorkingDir: wd,
			Prefix:     hashPrefix,
			ImageMap:   ImageMap{},
			HashLength: 32,
		})
		err := imgProcessor.ProcessAll(false)
		a.ErrorIs(err, ErrNoImages)
	})

	md := []MockImgProcData{
		{
			should:      "lowercase extension",
			files:       []string{"test1.PNG", "test2.JPG"},
			fileContent: []string{"test1", "test2"},
			expectFiles: []string{
				"0x@1b4f0e9851971998e732078544c96b36.png",
				"0x@60303ae22b998861bce3b28f33eec1be.jpg",
			},
		},
		{
			should: "ignore images with same hash",
			files: []string{
				"0x@1b4f0e9851971998e732078544c96b36.png",
				"test2.png",
				"test3.png",
				"test4.bmp",
			},
			fileContent: []string{"test1", "test2", "test3", "test4"},
			expectFiles: []string{
				"0x@1b4f0e9851971998e732078544c96b36.png",
				"0x@60303ae22b998861bce3b28f33eec1be.png",
				"0x@fd61a03af4f77d870fc21e05e7e80678.png",
				"0x@a4e624d686e03ed2767c0abd85c14426.bmp",
			},
		},
		{
			should: "delete duplicate images",
			files: []string{
				"test1.png",
				"test2.png",
				"test3.png",
				"test4.bmp",
				"test5.bmp",
				"test6.png",
				"test7.bmp",
				"test8.bmp",
			},
			fileContent: []string{
				"test1",
				"test1",
				"test1",
				"test4",
				"test4",
				"test1",
				"test4",
				"test4",
			},
			expectFiles: []string{
				"0x@1b4f0e9851971998e732078544c96b36.png",
				"0x@a4e624d686e03ed2767c0abd85c14426.bmp",
			},
		},
		{
			should: "handle large load of images",
			files: []string{
				"test1.bmp",
				"test2.png",
				"0x@fd61a03af4f77d870fc21e05e7e80678.png",
				"test4.bmp",
				"test5.png",
				"test6.bmp",
				"test7.bmp",
				"test8.bmp",
				"test9.bmp",
				"bad_file1.txt",
				"test10.png",
				"test11.bmp",
				"test12.bmp",
				"test13.bmp",
				"bad_file2.exe",
				"test14.bmp",
				"test15.bmp",
				"bad_file3.c",
			},
			fileContent: []string{
				"test3",
				"test2",
				"test3",
				"test4",
				"test2",
				"test3",
				"test7",
				"test8",
				"test9",
				"bad_file1",
				"test2",
				"test11",
				"test12",
				"test3",
				"bad_file2",
				"test14",
				"test15",
				"bad_file3",
			},
			expectFiles: []string{
				"0x@fd61a03af4f77d870fc21e05e7e80678.png",
				"0x@60303ae22b998861bce3b28f33eec1be.png",
				"0x@a4e624d686e03ed2767c0abd85c14426.bmp",
				"0x@bd7c911264aae15b66d4291b6850829a.bmp",
				"0x@1f9bfeb15fee8a10c4d0711c7eb0c083.bmp",
				"0x@b4451034d3b6590060ce9484a28b88dd.bmp",
				"0x@744ea9ec6fa0a83e9764b4e323d5be6b.bmp",
				"0x@a98ec5c5044800c88e862f007b98d898.bmp",
				"0x@40cca5cc13abf91c7d5a72c0aea9bcbe.bmp",
				"0x@ebb39b342baead7aa52c0bcd6c0d4ba0.bmp",
				"bad_file1.txt",
				"bad_file2.exe",
				"bad_file3.c",
			},
		},
	}

	for _, d := range md {
		t.Run("should "+d.should, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			dir := t.TempDir()
			err := writeFiles(dir, d.files, d.fileContent)
			a.NoError(err)

			fileNames, err := readDir(dir)
			a.NoError(err)
			a.Equal(
				len(d.fileContent),
				len(fileNames),
				"should always have the same number of files as file content",
			)
			iMap, err := MapImages(dir, hashPrefix)
			a.NoError(err)
			imgProcessor := NewImageProcessor(ImageProcessorConfig{
				WorkingDir: dir,
				Prefix:     hashPrefix,
				ImageMap:   iMap,
				HashLength: 32,
			})
			err = imgProcessor.ProcessAll(false)
			a.NoError(err)

			fileNames, err = readDir(dir)
			a.NoError(err)
			for _, fn := range fileNames {
				a.Contains(d.expectFiles, fn, "expected files should contain actual file")
			}

			a.Equal(len(d.expectFiles), len(fileNames))
		})
	}
}

func writeFiles(dir string, files []string, fileContent []string) error {
	for i, file := range files {
		err := os.WriteFile(dir+"/"+file, []byte(fileContent[i]), 0o644)
		if err != nil {
			return err
		}
	}
	return nil
}

func readDir(dir string) ([]string, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fileNames := []string{}
	for _, entry := range dirEntries {
		fileNames = append(fileNames, entry.Name())
	}
	return fileNames, nil
}
