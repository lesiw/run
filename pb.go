package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var root string

//go:embed version.txt
var version string

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (err error) {
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), "Usage of pb:\n\n")
		fmt.Fprint(flag.CommandLine.Output(), "    pb COMMAND [ARGS...]\n\n")
		flag.PrintDefaults()
	}
	var usermap stringlist
	flag.Var(&usermap, "u", "chowns files based on a given `mapping` uid:gid::uid:gid ")
	list := flag.Bool("l", false, "list all commands")
	printroot := flag.Bool("r", false, "print git root")
	install := flag.Bool("i", false, "install completion scripts")
	printversion := flag.Bool("v", false, "print version")
	flag.Parse()
	if *printversion {
		fmt.Print(version)
		return nil
	}

	if err = changeToGitRoot(); err != nil {
		return fmt.Errorf("could not find git root: %s", err)
	}
	if root, err = os.Getwd(); err != nil {
		return fmt.Errorf("could not get current working directory: %s", err)
	}
	if *list {
		return listCommands()
	} else if *printroot {
		fmt.Println(root)
		return nil
	} else if len(usermap) > 0 {
		return chownFiles(usermap)
	} else if *install {
		return installCompletion()
	}
	if flag.NArg() < 1 {
		return fmt.Errorf("no command given")
	}
	name := flag.Args()[0]
	var args []string
	if flag.NArg() > 1 {
		args = flag.Args()[1:]
	}
	cmdpath, err := findExecutable(name)
	if err != nil {
		return fmt.Errorf("error running command: %s", err)
	}
	cmd := exec.Command(cmdpath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("error running command: %s", err)
	}
	return nil
}

func changeToGitRoot() error {
	for {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		fileinfo, err := os.Stat(".git")
		if err == nil && fileinfo.IsDir() {
			return nil
		}
		reachedRoot := (cwd == "/" || cwd == (filepath.VolumeName(cwd)+"\\"))
		if reachedRoot || os.Chdir("..") != nil {
			return fmt.Errorf("No .git directory was found.")
		}
	}
}

func findExecutable(name string) (string, error) {
	oldpath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", oldpath) }()

	pbPath := pbPath()
	if pbPath != "" {
		path := pbPath + string(filepath.ListSeparator) + os.Getenv("PATH")
		if err := os.Setenv("PATH", path); err != nil {
			return "", fmt.Errorf("could not set PATH: %s", err)
		}
	}
	return exec.LookPath(name)
}

func pbPath() string {
	paths := os.Getenv("PBPATH")
	if paths == "" {
		paths = "./bin"
	} else if paths == "-" {
		return ""
	}
	abspaths := strings.Builder{}
	splitpaths := filepath.SplitList(paths)
	for i, path := range splitpaths {
		if i > 0 {
			abspaths.WriteString(string(filepath.ListSeparator))
		}
		parts := strings.Split(path, string(filepath.Separator))
		if len(parts) > 0 && parts[0] == "." {
			parts[0] = root
			abspaths.WriteString(strings.Join(parts, string(filepath.Separator)))
		} else {
			abspaths.WriteString(path)
		}
	}
	return abspaths.String()
}

func listCommands() error {
	paths, err := cmdPaths()
	if err != nil {
		return err
	}
	for _, path := range paths {
		fmt.Println(filepath.Base(path))
	}
	return nil
}

func cmdPaths() (cmds []string, err error) {
	paths := pbPath()
	var files []fs.DirEntry
	var info os.FileInfo
	for _, path := range filepath.SplitList(paths) {
		files, err = os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("error reading directory: %s", err)
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			info, err = file.Info()
			if err != nil {
				return
			}
			if info.Mode()&0111 != 0 {
				cmds = append(cmds, filepath.Join(path, file.Name()))
			}
		}
	}
	return
}

func chownFiles(mappings stringlist) error {
	for _, mapping := range mappings {
		fromstr, tostr, ok := strings.Cut(mapping, "::")
		if !ok {
			return fmt.Errorf("bad format: %s", mapping)
		}

		fuidstr, fgidstr, ok := strings.Cut(fromstr, ":")
		if !ok {
			return fmt.Errorf("bad user: %s", fromstr)
		}
		tuidstr, tgidstr, ok := strings.Cut(tostr, ":")
		if !ok {
			return fmt.Errorf("bad user: %s", tostr)
		}

		fuid, err := strconv.Atoi(fuidstr)
		if err != nil {
			return fmt.Errorf("bad id: %s", fromstr)
		}
		fgid, err := strconv.Atoi(fgidstr)
		if err != nil {
			return fmt.Errorf("bad id: %s", fromstr)
		}
		tuid, err := strconv.Atoi(tuidstr)
		if err != nil {
			return fmt.Errorf("bad id: %s", fromstr)
		}
		tgid, err := strconv.Atoi(tgidstr)
		if err != nil {
			return fmt.Errorf("bad id: %s", fromstr)
		}

		if err = chownDir(root, fuid, fgid, tuid, tgid); err != nil {
			return err
		}
	}
	return nil
}

type stringlist []string

func (s *stringlist) String() string {
	return "[" + strings.Join(*s, ", ") + "]"
}

func (s *stringlist) Set(v string) error {
	*s = append(*s, v)
	return nil
}
