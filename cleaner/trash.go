package cleaner

import (
	"errors"
	"fmt"
	"io"
	"mop/scanner"
	"os"
	"path/filepath"
	"syscall"
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

	if err := validateCleanPath(item.Path); err != nil {
		result.Success = false
		result.Error = err.Error()
		return result, err
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

	if err := moveToTrash(item.Path, destPath); err != nil {
		result.Success = false
		result.Error = err.Error()
		return result, err
	}

	result.Success = true
	return result, nil
}

// moveToTrash moves src to dst. It first tries an atomic rename, and falls back
// to a recursive copy + remove when src and dst live on different filesystems
// (os.Rename returns EXDEV in that case).
func moveToTrash(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}
	if !errors.Is(err, syscall.EXDEV) {
		return err
	}
	if cerr := copyRecursive(src, dst); cerr != nil {
		os.RemoveAll(dst)
		return cerr
	}
	return os.RemoveAll(src)
}

func copyRecursive(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dst)
	}
	if info.IsDir() {
		if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if err := copyRecursive(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	return copyFile(src, dst, info.Mode().Perm())
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
