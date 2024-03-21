//go:build windows
// +build windows

package main

import "fmt"

func chownDir(root string, fuid, fgid, tuid, tgid int) error {
	return fmt.Errorf("chownDir is not implemented on windows")
}
