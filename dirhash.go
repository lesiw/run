package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"
)

type hashfn func(files []string, fullpath func(string) string) (string, error)

// hashDir hashes the contents of path.
// The output of this function should be equivalent to:
//
//	for f in $(find . -type f | sort | cut -c 3-)
//	do
//	    if [ -x "$f" ]
//	    then
//	        printf "+ "
//	    else
//	        printf "  "
//	    fi
//	    sha256sum "$f"
//	done | sha256sum
func hash1(files []string, path func(string) string) (string, error) {
	h := sha256.New()
	files = append([]string{}, files...)
	sort.Strings(files)
	for _, file := range files {
		if strings.Contains(file, "\n") {
			return "", errors.New("files with newlines are not allowed")
		}
		r, err := os.Open(path(file))
		if err != nil {
			return "", err
		}
		hf := sha256.New()
		_, err = io.Copy(hf, r)
		r.Close()
		if err != nil {
			return "", err
		}
		info, err := os.Stat(path(file))
		if err != nil {
			return "", err
		}
		exec := ' '
		if info.Mode()&0111 != 0 {
			exec = '+'
		}
		fmt.Fprintf(h, "%c %x  %s\n", exec, hf.Sum(nil), file)
	}
	return "h1_" + hex.EncodeToString(h.Sum(nil)), nil
}

func hashDir(dir, prefix string, hash hashfn) (string, error) {
	files, err := dirhash.DirFiles(dir, prefix)
	if err != nil {
		return "", err
	}
	return hash(files, func(name string) string {
		return filepath.Join(dir, strings.TrimPrefix(name, prefix))
	})
}
