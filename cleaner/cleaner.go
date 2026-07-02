package cleaner

import "mop/scanner"

type CleanResult struct {
	Path    string
	Size    int64
	Success bool
	Error   string
}

type Cleaner interface {
	Clean(item scanner.CacheItem) (CleanResult, error)
}

type FileCleaner struct{}

func NewFileCleaner() *FileCleaner {
	return &FileCleaner{}
}
