package cleaner

import (
	"fmt"
	"mop/scanner"
	"os"
	"path/filepath"
	"time"
)

type TrashCleaner struct{}

func NewTrashCleaner() *TrashCleaner {
	return &TrashCleaner{}
}

func (c *TrashCleaner) Clean(item scanner.CacheItem) (CleanResult, error) {
	result := CleanResult{
		Path: item.Path,
		Size: item.Size,
	}

	home, err := os.UserHomeDir()
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result, err
	}

	trashDir := filepath.Join(home, ".Trash")
	if err := os.MkdirAll(trashDir, 0755); err != nil {
		result.Success = false
		result.Error = err.Error()
		return result, err
	}

	baseName := filepath.Base(item.Path)
	destName := baseName
	destPath := filepath.Join(trashDir, destName)

	if _, err := os.Stat(destPath); err == nil {
		timestamp := time.Now().Format("20060102-150405")
		ext := filepath.Ext(baseName)
		nameWithoutExt := baseName[:len(baseName)-len(ext)]
		destName = fmt.Sprintf("%s-%s%s", nameWithoutExt, timestamp, ext)
		destPath = filepath.Join(trashDir, destName)
	}

	err = os.Rename(item.Path, destPath)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result, err
	}

	result.Success = true
	return result, nil
}
