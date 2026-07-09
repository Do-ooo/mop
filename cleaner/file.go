package cleaner

import (
	"mop/scanner"
	"os"
)

func (f *FileCleaner) Clean(item scanner.CacheItem) (CleanResult, error) {
	result := CleanResult{
		Path: item.Path,
		Size: item.Size,
	}

	if err := validateCleanPath(item.Path); err != nil {
		result.Success = false
		result.Error = err.Error()
		return result, err
	}

	info, err := os.Stat(item.Path)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result, err
	}

	if info.IsDir() {
		err = os.RemoveAll(item.Path)
	} else {
		err = os.Remove(item.Path)
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result, err
	}

	result.Success = true
	return result, nil
}
