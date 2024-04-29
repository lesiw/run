package main

import (
	"maps"
	"os"
	"strings"
)

type runEnv struct {
	env  map[string]string
	argv []string
}

func (e *runEnv) Clone() *runEnv {
	o := &runEnv{}
	copy(o.argv, e.argv)
	for k, v := range e.env {
		o.env[k] = v
	}
	return o
}

// lpenv returns a copy of e.env that is compatible with lookpath.
func (e *runEnv) lpenv() map[string]string {
	m := make(map[string]string)
	maps.Copy(m, e.env)
	m["PATH"] = m["RUNPATH"]
	delete(m, "RUNPATH")
	return m
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
