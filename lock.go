package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (env *runEnv) LoadLocks() error {
	f, err := os.Open(filepath.Join(env.path, ".run", ".runlock"))
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for n := 1; scanner.Scan(); n++ {
		line := scanner.Text()
		url, rev, ok := strings.Cut(line, " ")
		if !ok {
			return fmt.Errorf("bad lock (line %d): '%s'", n, line)
		}
		env.locks[url] = rev
	}
	return nil
}

func (env *runEnv) SetLock(url, rev string) {
	if env.root == nil {
		env.locks[url] = rev
	} else {
		env.root.locks[url] = rev
	}
}

func (env *runEnv) WriteLocks() error {
	if err := os.MkdirAll(filepath.Join(env.path, ".run"), 0755); err != nil {
		return fmt.Errorf("failed to create .run directory: %w", err)
	}

	var lines []string
	for url, rev := range env.locks {
		lines = append(lines, fmt.Sprintf("%s %s", url, rev))
	}
	sort.Strings(lines)

	runlock := filepath.Join(env.path, ".run", ".runlock")
	err := os.WriteFile(runlock, []byte(strings.Join(lines, "\n")+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("failed to write .runlock: %w", err)
	}
	return nil
}
