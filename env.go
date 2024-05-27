package main

import (
	"maps"
	"os"
	"path/filepath"
	"strings"
)

type runEnv struct {
	env   map[string]string
	argv  []string
	locks map[string]string

	root *runEnv
	path string
}

func (e *runEnv) Clone() *runEnv {
	o := &runEnv{}
	o.path = e.path
	o.argv = append([]string{}, e.argv...)
	o.env = make(map[string]string)
	for k, v := range e.env {
		o.env[k] = v
	}
	o.locks = make(map[string]string)
	for k, v := range e.locks {
		o.locks[k] = v
	}
	if e.root == nil {
		o.root = e
	} else {
		o.root = e.root
	}
	return o
}

// lpenv returns a copy of e.env that is compatible with lookpath.
func (e *runEnv) lpenv() map[string]string {
	m := make(map[string]string)
	maps.Copy(m, e.env)
	var path strings.Builder
	parts := strings.Split(e.env["RUNPATH"], listsep)
	for _, part := range parts {
		if part == "" {
			continue
		}
		if path.Len() > 0 {
			path.WriteString(listsep)
		}
		path.WriteString(filepath.Join(part, ".run"))
	}
	delete(m, "RUNPATH")
	m["PATH"] = path.String()
	return m
}

func (env *runEnv) Apply() error {
	setenv(env.env)
	return env.WriteLocks()
}

func envmap() map[string]string {
	m := make(map[string]string)
	for _, env := range os.Environ() {
		k, v, _ := strings.Cut(env, "=")
		m[k] = v
	}
	return m
}

func setenv(env map[string]string) {
	for k, v := range env {
		os.Setenv(k, v)
	}
}

func baseEnv() *runEnv {
	return &runEnv{
		env:   envmap(),
		locks: make(map[string]string),
		path:  root,
	}
}
