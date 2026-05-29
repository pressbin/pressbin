package assets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Uploader interface {
	Upload(relPath string, data []byte) error
	Delete(relPath string) error
	Exists(relPath string) (bool, error)
}

type LocalUploader struct {
	StoragePath string
}

func NewLocalUploader(storagePath string) (*LocalUploader, error) {
	storagePath = strings.TrimSpace(storagePath)
	if storagePath == "" {
		return nil, fmt.Errorf("assets.path is not configured")
	}
	return &LocalUploader{StoragePath: storagePath}, nil
}

func (u *LocalUploader) fullPath(rel string) (string, error) {
	root := filepath.Clean(u.StoragePath)
	full := filepath.Join(root, filepath.FromSlash(rel))

	cleanRoot := filepath.Clean(root)
	cleanFull := filepath.Clean(full)
	if cleanFull != cleanRoot && !strings.HasPrefix(cleanFull, cleanRoot+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid asset path")
	}
	return cleanFull, nil
}

func (u *LocalUploader) Upload(relPath string, data []byte) error {
	full, err := u.fullPath(relPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}

func (u *LocalUploader) Delete(relPath string) error {
	full, err := u.fullPath(relPath)
	if err != nil {
		return err
	}
	return os.Remove(full)
}

func (u *LocalUploader) Exists(relPath string) (bool, error) {
	full, err := u.fullPath(relPath)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(full)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

