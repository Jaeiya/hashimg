package lib

import "errors"

var (
	ErrNoImages = errors.New("no images found in directory")

	ErrHashPrefixTooShort = errors.New("hash prefix must be at least 3 characters")
	ErrHashInfoNil        = errors.New("hash info is nil; it must be initialized")
	ErrHashLengthTooShort = errors.New("hash length must be at least 10 characters")
)
