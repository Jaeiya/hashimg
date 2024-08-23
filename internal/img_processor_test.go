package internal

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockImgProcData struct {
	should      string
	files       []string
	fileContent []string
	expectFiles []string

	expectDupeCount      int32
	expectUpdateProgress int32

	//###########################
	// The following are review process fields
	//###########################

	dupeFiles        []string
	dupeContent      []string
	expectDupeFiles  []string
	expectImageCount int32

	// Should only tally new images because dupes are handled by the
	// restoration method.
	expectMaxUpdateProgress int32
}

const hashLength = 32

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
			HashLength: hashLength,
		})
		err := imgProcessor.ProcessAll(false)
		a.ErrorIs(err, ErrNoImages)
	})

	md := []MockImgProcData{
		{
			should:               "lowercase extension",
			files:                []string{"test1.PNG", "test2.JPG"},
			fileContent:          []string{"test1", "test2"},
			expectUpdateProgress: 2,
			expectDupeCount:      0,
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
			fileContent:          []string{"test1", "test2", "test3", "test4"},
			expectUpdateProgress: 3,
			expectDupeCount:      0,
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
			expectUpdateProgress: 8,
			expectDupeCount:      6,
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
			expectUpdateProgress: 14,
			expectDupeCount:      5,
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
				HashLength: hashLength,
			})
			err = imgProcessor.ProcessAll(false)
			a.NoError(err)

			a.Equal(
				d.expectDupeCount,
				imgProcessor.Status.DupeImageCount,
				"dupe count should be reliable",
			)

			a.Equal(
				d.expectUpdateProgress,
				imgProcessor.Status.UpdateProgress,
				"update progress should be reliable",
			)
			a.Equal(
				d.expectUpdateProgress,
				imgProcessor.Status.MaxUpdateProgress,
				"total update progress should always equal max",
			)

			fileNames, err = readDir(dir)
			a.NoError(err)
			for _, fn := range fileNames {
				a.Contains(d.expectFiles, fn, "expected files should contain actual file")
			}

			a.Equal(len(d.expectFiles), len(fileNames))
		})
	}
}

func TestReviewProcess(t *testing.T) {
	hashPrefix := "0x@"

	md := []MockImgProcData{
		{
			should:                  "move duplicate new files",
			dupeFiles:               []string{"t0.jpg", "t1.jpg"},
			dupeContent:             []string{"0", "0"},
			expectDupeCount:         1,
			expectImageCount:        1,
			expectMaxUpdateProgress: 1,
			expectDupeFiles: []string{
				fmt.Sprintf("%s_1.jpg", calcSha256("0")),
				fmt.Sprintf("%s_2.jpg", calcSha256("0")),
			},
		},
		{
			should: "move lots of duplicate new images",
			dupeFiles: []string{
				"t0.jpg",
				"t1.jpg",
				"t2.jpg",
				"t3.jpg",
				"t4.jpg",
				"t5.jpg",
				"t6.jpg",
			},
			dupeContent:             []string{"0", "0", "0", "0", "1", "1", "1"},
			expectImageCount:        2,
			expectDupeCount:         5,
			expectMaxUpdateProgress: 2,
			expectDupeFiles: []string{
				fmt.Sprintf("%s_1.jpg", calcSha256("0")),
				fmt.Sprintf("%s_2.jpg", calcSha256("0")),
				fmt.Sprintf("%s_3.jpg", calcSha256("0")),
				fmt.Sprintf("%s_4.jpg", calcSha256("0")),
				fmt.Sprintf("%s_1.jpg", calcSha256("1")),
				fmt.Sprintf("%s_2.jpg", calcSha256("1")),
				fmt.Sprintf("%s_3.jpg", calcSha256("1")),
			},
		},
		{
			should: "move duplicate cached files",
			dupeFiles: []string{
				fmt.Sprintf("0x@%s.jpg", calcSha256("0")),
				"t0.jpg",
			},
			expectImageCount:        0,
			expectDupeCount:         1,
			expectMaxUpdateProgress: 0,
			dupeContent:             []string{"0", "0"},
			expectDupeFiles: []string{
				fmt.Sprintf("%s_1.jpg", calcSha256("0")),
				fmt.Sprintf("%s_2.jpg", calcSha256("0")),
			},
		},
		{
			should: "move combination of new and cached duplicate files",
			dupeFiles: []string{
				fmt.Sprintf("0x@%s.jpg", calcSha256("0")),
				fmt.Sprintf("0x@%s.jpg", calcSha256("1")),
				fmt.Sprintf("0x@%s.jpg", calcSha256("2")),
				"t1.jpg",
				"t2.jpg",
				"t3.jpg",
				"t4.jpg",
				"t5.jpg",
				"t6.jpg",
				"t7.jpg",
				"t8.jpg",
				"t9.jpg",
				"t10.jpg",
				"t11.jpg",
				"t12.jpg",
				"t13.jpg",
				"t14.jpg",
			},
			dupeContent: []string{
				"0",
				"1",
				"2",
				"0",
				"0",
				"0",
				"0",
				"1",
				"1",
				"1",
				"2",
				"2",
				"2",
				"2",
				"3",
				"3",
				"3",
			},
			expectImageCount:        1,
			expectDupeCount:         13,
			expectMaxUpdateProgress: 1,
			expectDupeFiles: []string{
				fmt.Sprintf("%s_1.jpg", calcSha256("0")),
				fmt.Sprintf("%s_2.jpg", calcSha256("0")),
				fmt.Sprintf("%s_3.jpg", calcSha256("0")),
				fmt.Sprintf("%s_4.jpg", calcSha256("0")),
				fmt.Sprintf("%s_5.jpg", calcSha256("0")),
				fmt.Sprintf("%s_1.jpg", calcSha256("1")),
				fmt.Sprintf("%s_2.jpg", calcSha256("1")),
				fmt.Sprintf("%s_3.jpg", calcSha256("1")),
				fmt.Sprintf("%s_4.jpg", calcSha256("1")),
				fmt.Sprintf("%s_1.jpg", calcSha256("2")),
				fmt.Sprintf("%s_2.jpg", calcSha256("2")),
				fmt.Sprintf("%s_3.jpg", calcSha256("2")),
				fmt.Sprintf("%s_4.jpg", calcSha256("2")),
				fmt.Sprintf("%s_5.jpg", calcSha256("2")),
				fmt.Sprintf("%s_1.jpg", calcSha256("3")),
				fmt.Sprintf("%s_2.jpg", calcSha256("3")),
				fmt.Sprintf("%s_3.jpg", calcSha256("3")),
			},
		},
		{
			should: "leave non-dupe files alone",
			files: []string{
				fmt.Sprintf("0x@%s.jpg", calcSha256("0")),
				fmt.Sprintf("0x@%s.jpg", calcSha256("1")),
				fmt.Sprintf("0x@%s.jpg", calcSha256("2")),
				"t5.jpg",
				"t6.jpg",
			},
			fileContent: []string{
				"0",
				"1",
				"2",
				"5",
				"6",
			},
			expectFiles: []string{
				fmt.Sprintf("0x@%s.jpg", calcSha256("0")),
				fmt.Sprintf("0x@%s.jpg", calcSha256("1")),
				fmt.Sprintf("0x@%s.jpg", calcSha256("2")),
				"t5.jpg",
				"t6.jpg",
			},
			dupeFiles: []string{
				fmt.Sprintf("0x@%s.jpg", calcSha256("3")),
				"t0.jpg",
				"t1.jpg",
				"t2.jpg",
				"t3.jpg",
			},
			dupeContent: []string{
				"3",
				"3",
				"3",
				"4",
				"4",
			},
			expectImageCount:        3,
			expectDupeCount:         3,
			expectMaxUpdateProgress: 3,
			expectDupeFiles: []string{
				fmt.Sprintf("%s_1.jpg", calcSha256("3")),
				fmt.Sprintf("%s_2.jpg", calcSha256("3")),
				fmt.Sprintf("%s_3.jpg", calcSha256("3")),
				fmt.Sprintf("%s_1.jpg", calcSha256("4")),
				fmt.Sprintf("%s_2.jpg", calcSha256("4")),
			},
		},
		{
			should: "do nothing if no duplicate images",
			files: []string{
				fmt.Sprintf("0x@%s.jpg", calcSha256("0")),
				fmt.Sprintf("0x@%s.jpg", calcSha256("1")),
				fmt.Sprintf("0x@%s.jpg", calcSha256("2")),
			},
			fileContent: []string{
				"0",
				"1",
				"2",
			},
			expectImageCount:        0,
			expectDupeCount:         0,
			expectMaxUpdateProgress: 0,
			expectFiles: []string{
				fmt.Sprintf("0x@%s.jpg", calcSha256("0")),
				fmt.Sprintf("0x@%s.jpg", calcSha256("1")),
				fmt.Sprintf("0x@%s.jpg", calcSha256("2")),
			},
		},
	}
	_ = md

	for _, d := range md {
		t.Run("should "+d.should, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			dir := t.TempDir()
			hasDupes := len(d.dupeFiles) > 0
			hasFiles := len(d.files) > 0

			if hasFiles {
				err := writeFiles(dir, d.files, d.fileContent)
				require.NoError(t, err)
			}

			if hasDupes {
				err := writeFiles(dir, d.dupeFiles, d.dupeContent)
				require.NoError(t, err)
			}

			iMap, err := MapImages(dir, hashPrefix)
			a.NoError(err)
			imgProcessor := NewImageProcessor(ImageProcessorConfig{
				WorkingDir: dir,
				Prefix:     hashPrefix,
				ImageMap:   iMap,
				HashLength: hashLength,
			})

			err = imgProcessor.ProcessHashReview(false)
			require.NoError(t, err, "process hashes and move files without error")

			// The dupes are handled by the restoration method
			a.Equal(
				d.expectMaxUpdateProgress,
				imgProcessor.Status.NewImageCount,
				"total new images should match the max update progress",
			)

			if hasDupes {
				_, err = os.Stat(filepath.Join(dir, "__dupes"))
				require.NoError(t, err, "dupes folder should exist")
				fileNames, err := readDir(filepath.Join(dir, "__dupes"))
				require.NoError(t, err, "read directory without error")
				for _, fn := range fileNames {
					a.Contains(
						d.expectDupeFiles,
						fn,
						"expected files should contain actual file",
					)
				}
				a.Len(d.expectDupeFiles, len(fileNames))
				a.Equal(
					d.expectDupeCount,
					imgProcessor.Status.DupeImageCount,
					"should have accurate dupe count",
				)
			}

			if !hasDupes {
				_, err = os.Stat(filepath.Join(dir, "__dupes"))
				a.Error(err, "temp dupes dir should not exist")
			}

			if len(d.files) > 0 {
				fileNames, err := readDir(dir)
				require.NoError(t, err, "read directory without error")

				// If dupes exist then we don't count __dupes dir
				adjustedLen := 1
				if !hasDupes {
					adjustedLen -= 1
				}
				require.Len(
					t,
					fileNames,
					len(d.files)+adjustedLen,
					"we should have files that were not moved",
				)
				for _, fn := range fileNames {
					// Ignore directory name
					if fn == "__dupes" {
						continue
					}
					a.Contains(d.expectFiles, fn, "expect non-duplicates to be moved")
				}
			}

			a.Equal(
				d.expectImageCount,
				imgProcessor.Status.NewImageCount,
				"should have accurate image count",
			)

			if len(d.files) == 0 {
				files, err := readDir(dir)
				require.NoError(t, err)
				a.Len(files, 1, "all files should be moved to dupes folder")
			}
		})
	}

}

func TestReviewRestoration(t *testing.T) {
	hashPrefix := "0x@"

	t.Run("should restore new images and delete temp folder", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		dir := t.TempDir()
		writeFiles(
			dir,
			[]string{
				"t1.jpg",
				"t2.jpg",
				"t3.jpg",
				"t4.jpg",
				fmt.Sprintf("0x@%s.jpg", calcSha256("2")),
				"t6.jpg",
				"t7.jpg",
			},
			[]string{"0", "0", "1", "1", "2", "2", "2"},
		)

		iMap, err := MapImages(dir, hashPrefix)
		a.NoError(err)
		imgProcessor := NewImageProcessor(ImageProcessorConfig{
			WorkingDir: dir,
			Prefix:     hashPrefix,
			ImageMap:   iMap,
			HashLength: hashLength,
		})

		err = imgProcessor.ProcessHashReview(false)
		require.NoError(t, err)

		err = imgProcessor.RestoreFromReview()
		require.NoError(t, err)

		fileNames, err := readDir(dir)
		require.NoError(t, err)
		a.Len(fileNames, 3, "there should only be 3 preserved files")
		a.NotContains(fileNames, "__dupes", "temp dupe folder should not exist")
		a.Contains(fileNames, fmt.Sprintf("%s_1.jpg", calcSha256("0")))
		a.Contains(fileNames, fmt.Sprintf("%s_1.jpg", calcSha256("1")))
		a.Contains(fileNames, fmt.Sprintf("%s_1.jpg", calcSha256("2")))
	})

}

func writeFiles(dir string, files []string, fileContent []string) error {
	if len(files) != len(fileContent) {
		return fmt.Errorf("files length does not match file content length")
	}
	for i, file := range files {
		err := os.WriteFile(dir+"/"+file, []byte(fileContent[i]), 0o644)
		if err != nil {
			return err
		}
	}
	return nil
}

func calcSha256(v string) string {
	s := sha256.New()
	s.Write([]byte(v))
	return fmt.Sprintf("%x", s.Sum(nil))[:hashLength]
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
