package lib

type HashResult struct {
	hash string
	path string
	// If image already renamed to its hash with prefix
	cached bool
	err    error
}

type FilterResult struct {
	newImageHashes  map[string]string
	dupeImageHashes []string
}
