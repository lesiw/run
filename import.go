package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	bysrc, err := cacheDir("store", "by-src")
	if err != nil {
		return "", err
	}
	cachepath := filepath.Join(bysrc, filepath.Base(src))
	if _, err := os.Lstat(cachepath); err == nil {
		if cachepath, err := filepath.EvalSymlinks(cachepath); err != nil {
			return "", fmt.Errorf("failed evaluating symlink: %w", err)
		} else {
			return cachepath, nil
		}
	}

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
	hash, err := hashDir(out, "", hash1)
	if err != nil {
		return "", err
	}
	cache, err := cacheDir("store")
	if err != nil {
		return "", err
	}
	path := filepath.Join(cache, hash)
	if err = storeCopy(out, path); err != nil {
		return "", fmt.Errorf("failed copying output dir: %w", err)
	}
	relpath, err := filepath.Rel(bysrc, path)
	if err != nil {
		return "", fmt.Errorf(
			"failed calculating relative path to package output: %w", err)
	}
	err = os.Symlink(relpath, cachepath)
	if err != nil {
		return "", fmt.Errorf("failed creating by-src symlink: %w", err)
	}
	return path, nil
}

func storeCopy(src, dst string) error {
	copyfunc := func(srcpath string, d fs.DirEntry, _ error) error {
		relpath, err := filepath.Rel(src, srcpath)
		if err != nil {
			return err
		}
		dstpath := filepath.Join(dst, relpath)
		if d.IsDir() {
			return os.Mkdir(dstpath, 0777)
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("failed to copy '%s': not a regular file",
				srcpath)
		}
		info, err := os.Stat(srcpath)
		if err != nil {
			return err
		}
		var exec fs.FileMode
		if info.Mode()&0111 != 0 {
			exec = 0111
		}
		srcfile, err := os.OpenFile(srcpath, os.O_RDONLY, 0)
		if err != nil {
			return err
		}
		defer srcfile.Close()
		dstfile, err := os.OpenFile(dstpath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666|exec)
		if err != nil {
			return err
		}
		defer dstfile.Close()
		if _, err = io.Copy(dstfile, srcfile); err != nil {
			return err
		}
		if err = os.Chmod(dstpath, 0444|exec); err != nil {
			return err
		}
		return nil
	}
	return filepath.WalkDir(src, copyfunc)
}
