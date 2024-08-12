package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasher(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	t.Run("should error if hash prefix is too short", func(t *testing.T) {
		t.Parallel()
		_, err := NewHasher(HasherConfig{
			Length:    32,
			Threads:   2,
			QueueSize: 100,
			Prefix:    "t",
		})
		a.ErrorIs(err, ErrHashPrefixTooShort)
	})

	t.Run("should error if hash info is nil", func(t *testing.T) {
		t.Parallel()
		_, err := NewHasher(HasherConfig{
			Length:     32,
			Threads:    2,
			QueueSize:  100,
			HashResult: nil,
			Prefix:     "tst",
		})
		a.ErrorIs(err, ErrHashInfoNil)
	})

	t.Run("should error if hash length is too short", func(t *testing.T) {
		t.Parallel()
		_, err := NewHasher(HasherConfig{
			Length:     9,
			Threads:    2,
			QueueSize:  100,
			HashResult: &HashResult{},
			Prefix:     "tst",
		})
		a.ErrorIs(err, ErrHashLengthTooShort)
	})
}
