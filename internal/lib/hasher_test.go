package lib

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasher(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	_ = a

	h := MyHasher{}

	testCases := []struct {
		path      string
		r         io.Reader
		len       int
		getActual func() string
	}{
		{
			path: "",
			r:    strings.NewReader("Hello World"),
			len:  24,
			getActual: func() string {
				fixedLen := "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e"[:24]
				return fixedLen
			},
		},
		{
			path: "",
			r:    strings.NewReader("This is some more test text"),
			len:  10,
			getActual: func() string {
				fixedLen := "dabc55d5d56fef58e477d00fe2a163785a1a517c918a4edc01e48198ff14b341"[:10]
				return fixedLen
			},
		},
		{
			path: "",
			r:    strings.NewReader("Simulate some more data"),
			len:  50,
			getActual: func() string {
				fixedLen := "f2ded4da50355624f36acc01db158fa0ad4a13ec15c27ea3d3673f2b249ad016"[:50]
				return fixedLen
			},
		},
		{
			path: "",
			r:    strings.NewReader("1234567890"),
			len:  64,
			getActual: func() string {
				fixedLen := "c775e7b757ede630cd0aa1113bd102661ab38829ca52a6422ab782862f268646"[:64]
				return fixedLen
			},
		},
	}

	for _, tc := range testCases {
		r := h.Hash(tc.r, tc.path, tc.len)
		a.Equal(r.hash, tc.getActual(), "should match length and characters of expected hash")
	}
}
