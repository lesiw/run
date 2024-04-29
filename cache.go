package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func cacheDir(path string) (cache string, err error) {
	if cache, err = os.UserCacheDir(); err != nil {
		return "", fmt.Errorf("failed to get user cache directory: %s", err)
	}
	cache = filepath.Join(cache, "run", path)
	if err = os.MkdirAll(cache, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %s", err)
	}
	return
}
