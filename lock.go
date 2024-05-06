package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func lock(url string) string {
	f, err := os.Open(filepath.Join(".run", ".runlock"))
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if u, rev, _ := strings.Cut(scanner.Text(), " "); u == url {
			return rev
		}
	}
	return ""
}

func setLock(url, rev string) error {
	if err := os.MkdirAll(".run", 0755); err != nil {
		return fmt.Errorf("failed to create .run directory: %w", err)
	}
	line := fmt.Sprintf("%s %s", url, rev)

	runlock := filepath.Join(".run", ".runlock")
	if _, err := os.Stat(runlock); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat .runlock: %w", err)
		}
		if err = os.WriteFile(runlock, []byte(line+"\n"), 0644); err != nil {
			return fmt.Errorf("failed to write .runlock: %w", err)
		}
		return nil
	}

	f, err := os.Open(runlock)
	if err != nil {
		return fmt.Errorf("failed to read .runlock: %w", err)
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()
		if u, _, _ := strings.Cut(l, " "); u != url {
			lines = append(lines, l)
		}
	}
	lines = append(lines, line)
	f.Close()
	f = nil

	sort.Strings(lines)
	err = os.WriteFile(runlock, []byte(strings.Join(lines, "\n")+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("failed to write .runlock: %w", err)
	}
	return nil
}
