package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const rundlurl = "https://github.com/lesiw/run/releases"

var unameos = map[string]string{
	"linux": "Linux",
}

var unamearch = map[string]string{
	"386":   "i386",
	"amd64": "x86_64",
	"arm":   "armv7l",
	"arm64": "aarch64",
}

func fetchRun(binos, arch string) (string, error) {
	if binos == runtime.GOOS && arch == runtime.GOARCH {
		path, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf(
				"failed to find path of current executable: %s", err)
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
	url := rundlurl + "/download/" + version + "/run-" + urlos + "-" + urlarch
	cache, err := cacheDir("bin")
	if err != nil {
		return "", err
	}
	path := filepath.Join(cache, "run-"+version+"-"+binos+"-"+arch)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	if err = downloadUrl(url, path); err != nil {
		return "", err
	}
	return path, nil
}

func downloadUrl(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch url '%s': %s", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: http status %d",
			resp.StatusCode)
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %s", path, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to download to '%s': %s", path, err)
	}

	if err = os.Chmod(path, 0755); err != nil {
		return fmt.Errorf("failed to mark '%s' as executable: %s", path, err)
	}

	return nil
}
