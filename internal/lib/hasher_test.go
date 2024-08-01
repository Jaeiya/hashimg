package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasher(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	t.Run("should error if hash length is too small", func(t *testing.T) {
		t.Parallel()
		_, err := NewHasher(HasherConfig{
			Length:    1,
			Threads:   10,
			QueueSize: 100,
		})
		a.ErrorIs(err, ErrHashPrefixTooShort)
	})

	t.Run("should error if hash info is nil", func(t *testing.T) {
		t.Parallel()
		_, err := NewHasher(HasherConfig{
			Length:    32,
			Threads:   10,
			QueueSize: 100,
			HashInfo:  nil,
			Prefix:    "test",
		})
		a.ErrorIs(err, ErrHashInfoNil)
	})
}
