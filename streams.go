package main

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"
)

type streamLogger struct {
	strings.Builder
}
type prefixWriter struct {
	p []byte
	w io.Writer
}

var lastlog streamLogger

func captureCmdUnlessVerbose() *exec.Cmd {
	if *verbose {
		return stdCmd()
	} else {
		return logCmd()
	}
}

func stdCmd() *exec.Cmd {
	return &exec.Cmd{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

func logCmd() *exec.Cmd {
	lastlog.Reset()
	if isTty() {
		return &exec.Cmd{
			Stdout: newPrefixWriter("\033[0m", &lastlog),
			Stderr: newPrefixWriter("\033[31m", &lastlog),
		}
	} else {
		return &exec.Cmd{
			Stdout: &lastlog,
			Stderr: &lastlog,
		}
	}
}

func newPrefixWriter(prefix string, writer io.Writer) *prefixWriter {
	return &prefixWriter{p: []byte(prefix), w: writer}
}

func (p *prefixWriter) Write(b []byte) (n int, err error) {
	_, _ = p.w.Write(p.p)
	return p.w.Write(b)
}

func (s *streamLogger) String() string {
	return s.Builder.String() + "\033[0m"
}

func isTty() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
