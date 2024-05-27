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

func getPackage(env *runEnv, url string) error {
	path, err := packageOut(env, url)
	if err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}

func importPackage(env *runEnv, url string) error {
	path, err := packageOut(env, url)
	if err != nil {
		return err
	}
	bin := filepath.Join(path, ".run")
	if _, err = os.Stat(bin); err != nil {
		return fmt.Errorf("failed to import '%s': no .run directory", url)
	}
	env.env["RUNPATH"] = env.env["RUNPATH"] + listsep + path + listsep
	return nil
}

func packageOut(env *runEnv, url string) (out string, err error) {
	rev := env.locks[url]
	if !strings.Contains(url, "@") {
		if !strings.Contains(url, "://") {
			url = "https://" + url
		}
		url, err = realUrl(url)
		if err != nil {
			return "", fmt.Errorf("failed to fetch url '%s': %w", url, err)
		}
	}
	if rev != "" {
		cachepath, err := storeByRev(rev)
		if err != nil {
			return "", err
		}
		if cachepath != "" {
			return cachepath, nil
		}
	}
	src, rev, err := packageSrc(url, rev)
	if err != nil {
		return "", err
	}
	out, err = packageBuild(src)
	if err != nil {
		return "", fmt.Errorf("failed to build '%s': %w", url, err)
	}
	env.SetLock(url, rev)
	return out, nil
}

func packageSrc(url, rev string) (string, string, error) {
	dir, err := os.MkdirTemp("", "run")
	if err != nil {
		return "", rev,
			fmt.Errorf("failed to create temp directory: %w", err)
	}
	defers.add(func() { _ = os.RemoveAll(dir) })
	cmd := exec.Command("git", "clone", url, dir)
	if *verbose {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		// TODO: if not verbose, return stderr
		return "", rev, fmt.Errorf("failed to clone '%s': %w", url, err)
	}

	if rev == "" {
		cmd = exec.Command("git", "-C", dir, "rev-parse", "HEAD")
		buf, err := cmd.Output()
		if err != nil {
			return "", rev,
				fmt.Errorf("failed to get HEAD rev from '%s': %w", url, err)
		}
		rev = strings.Trim(string(buf), "\n")
	} else {
		cmd = exec.Command("git", "-C", dir, "checkout", rev)
		if *verbose {
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
		}
		if err := cmd.Run(); err != nil {
			// TODO: if not verbose, return stderr
			return "", rev, fmt.Errorf(
				"failed to checkout rev '%s' from '%s': %w", rev, url, err)
		}
	}

	cache, err := cacheDir("src")
	if err != nil {
		return "", rev, err
	}

	path := filepath.Join(cache, rev)
	if _, err = os.Stat(path); err == nil {
		return path, rev, nil
	}

	err = os.Rename(dir, path)
	return path, rev, err
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

func storeByRev(rev string) (string, error) {
	bysrc, err := cacheDir("store", "by-src")
	if err != nil {
		return "", err
	}
	cachepath := filepath.Join(bysrc, rev)
	if _, err := os.Lstat(cachepath); err == nil {
		if cachepath, err := filepath.EvalSymlinks(cachepath); err != nil {
			return "", fmt.Errorf("failed evaluating symlink: %w", err)
		} else {
			return cachepath, nil
		}
	}
	return "", nil
}
