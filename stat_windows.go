//go:build windows
// +build windows

package main

import "fmt"

func chownDir(string, int, int, int, int) error {
	return fmt.Errorf("chownDir is not implemented for windows")
}

func getOwner(string) (int, int, error) {
	return 0, 0, fmt.Errorf("getOwner is not implemented for windows")
}
