//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

func chownDir(root string, fuid, fgid, tuid, tgid int) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get info for %s: %s", d.Name(), err)
		}
		if int(stat.Uid) != fuid || int(stat.Gid) != fgid {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if err := os.Chown(path, tuid, tgid); err != nil {
			return fmt.Errorf("failed to chown %s: %s", d.Name(), err)
		}
		return nil
	})
}

func getOwner(path string) (uid int, gid int, err error) {
	info, err := os.Lstat(path)
	if err != nil {
		return
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		err = fmt.Errorf("failed to get info for %s: %s", path, err)
		return
	}
	uid = int(stat.Uid)
	gid = int(stat.Gid)
	return
}

func getMtime(path string) (mtime int64, err error) {
	var info fs.FileInfo
	info, err = os.Lstat(path)
	if err != nil {
		return
	}
	mtime = info.ModTime().Unix()
	return
}
