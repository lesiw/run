package main

import (
	"fmt"
	"os"

	lua "github.com/yuin/gopher-lua"
)

func runInit(e *runEnv, p string) error {
	if _, err := os.Stat(p); err != nil {
		return nil
	}
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer L.Close()

	cfg, argt, envt := L.NewTable(), L.NewTable(), L.NewTable()
	for _, arg := range e.argv {
		argt.Append(lua.LString(arg))
	}
	L.SetField(cfg, "argv", argt)
	for k, v := range e.env {
		L.SetField(envt, k, lua.LString(v))
	}
	L.SetField(cfg, "env", envt)
	L.SetGlobal("run", cfg)

	if err := L.DoFile(p); err != nil {
		return fmt.Errorf("failed to run init.lua: %s", err)
	}

	e.argv = []string{}
	argt.ForEach(func(_, v lua.LValue) { e.argv = append(e.argv, v.String()) })
	for k := range e.env {
		delete(e.env, k)
	}
	envt.ForEach(func(k, v lua.LValue) { e.env[k.String()] = v.String() })
	return nil
}
