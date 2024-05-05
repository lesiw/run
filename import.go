package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	cp "github.com/otiai10/copy"
	"golang.org/x/mod/sumdb/dirhash"
)

func getPackage(url string) error {
	path, err := packageOut(url)
	if err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}

func importPackage(url string) error {
	path, err := packageOut(url)
	if err != nil {
		return err
	}
	bin := filepath.Join(path, ".run")
	if _, err = os.Stat(bin); err != nil {
		return fmt.Errorf("failed to import '%s': no .run directory", url)
	}
	err = os.Setenv("RUNPATH", os.Getenv("RUNPATH")+listsep+path)
	if err != nil {
		return fmt.Errorf("failed to set RUNPATH: %s", err)
	}
	return nil
}

func packageOut(url string) (out string, err error) {
	if !strings.Contains(url, "@") {
		if !strings.Contains(url, "://") {
			url = "https://" + url
		}
		url, err = realUrl(url)
		if err != nil {
			return "", fmt.Errorf("failed to fetch url '%s': %w", url, err)
		}
	}
	src, err := packageSrc(url)
	if err != nil {
		return "", err
	}
	out, err = packageBuild(src)
	if err != nil {
		return "", fmt.Errorf("failed to build '%s': %w", url, err)
	}
	return out, nil
}

func packageSrc(url string) (string, error) {
	dir, err := os.MkdirTemp("", "run")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defers.add(func() { _ = os.RemoveAll(dir) })
	cmd := exec.Command("git", "clone", "--depth=1", url, dir)
	if *verbose {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to clone '%s': %w", url, err)
	}

	cmd = exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	buf, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD rev from '%s': %w", url, err)
	}
	sha := strings.Trim(string(buf), "\n")

	cache, err := cacheDir("src")
	if err != nil {
		return "", err
	}

	path := filepath.Join(cache, sha)
	if _, err = os.Stat(path); err == nil {
		return path, nil
	}

	err = os.Rename(dir, path)
	return path, err
}

func packageBuild(src string) (string, error) {
	run, err := os.Executable()
	if err != nil {
		run = "run"
	}
	cmd := exec.Command(run)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = src
	if err = cmd.Run(); err != nil {
		return "", fmt.Errorf("run failed: %w", err)
	}
	out := filepath.Join(src, "out")
	if _, err = os.Stat(out); err != nil {
		return "", err
	}
	hash, err := hashDir(out)
	if err != nil {
		return "", err
	}
	cache, err := cacheDir("var", "pkg")
	if err != nil {
		return "", err
	}
	path := filepath.Join(cache, hash)
	if err = cp.Copy(out, path); err != nil {
		return "", fmt.Errorf("failed copying output dir: %w", err)
	}
	return path, nil
}

// hashDir hashes the contents of path.
// The output of this function should be equivalent to:
//
//	sha256sum $(find . -type f | sort | cut -c 3-) | sha256sum
func hashDir(path string) (string, error) {
	h1, err := dirhash.HashDir(path, "", dirhash.Hash1)
	if err != nil {
		return "", err
	}
	kind, b64hash, ok := strings.Cut(h1, ":")
	if !ok || kind != "h1" {
		return "", fmt.Errorf("bad hash format from dirhash: %s", h1)
	}
	rawhash, err := base64.StdEncoding.DecodeString(b64hash)
	if err != nil {
		return "", fmt.Errorf("bad base64 string from dirhash: %s", b64hash)
	}
	return kind + "_" + hex.EncodeToString(rawhash), nil
}
