package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const pbdlurl = "https://github.com/lesiw/pb/releases"

var unameos = map[string]string{
	"linux": "Linux",
}

var unamearch = map[string]string{
	"386":   "i386",
	"amd64": "x86_64",
	"arm":   "armv7l",
	"arm64": "aarch64",
}

func fetchPb(binos, arch string) (string, error) {
	if binos == runtime.GOOS && arch == runtime.GOARCH {
		path, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("could not find path of current executable: %s", err)
		}
		return path, nil
	}
	urlos, ok := unameos[binos]
	if !ok {
		return "", fmt.Errorf("unrecognized container os: %s", binos)
	}
	urlarch, ok := unamearch[arch]
	if !ok {
		return "", fmt.Errorf("unrecognized container arch: %s", arch)
	}
	url := pbdlurl + "/download/" + version + "/pb-" + urlos + "-" + urlarch
	cache, err := cacheDir()
	if err != nil {
		return "", err
	}
	bincache := filepath.Join(cache, "bin")
	if err = os.MkdirAll(bincache, 0755); err != nil {
		return "", fmt.Errorf("could not create cache directory '%s': %s", bincache, err)
	}
	path := filepath.Join(bincache, "pb-"+version+"-"+binos+"-"+arch)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	if err = downloadUrl(url, path); err != nil {
		return "", err
	}
	return path, nil
}

func cacheDir() (cache string, err error) {
	if cache, err = os.UserCacheDir(); err != nil {
		return "", fmt.Errorf("could not get user cache directory: %s", err)
	}
	cache = filepath.Join(cache, "pb")
	if err = os.MkdirAll(cache, 0755); err != nil {
		return "", fmt.Errorf("could not create cache directory: %s", err)
	}
	return
}

func downloadUrl(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("could not fetch url '%s': %s", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not download file: http status %d", resp.StatusCode)
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create file '%s': %s", path, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("could not download to '%s': %s", path, err)
	}

	if err = os.Chmod(path, 0755); err != nil {
		return fmt.Errorf("could not mark '%s' as executable: %s", path, err)
	}

	return nil
}
