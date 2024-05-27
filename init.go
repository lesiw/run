package main

import (
	"fmt"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

func (env *runEnv) Init() error {
	env.env["RUNPATH"] = env.path
	if err := env.LoadLocks(); err != nil {
		return err
	}
	script := filepath.Join(env.path, ".run", "init.lua")
	if _, err := os.Stat(script); err != nil {
		return nil
	}
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer L.Close()

	cfg, argt, envt := L.NewTable(), L.NewTable(), L.NewTable()
	for _, arg := range env.argv {
		argt.Append(lua.LString(arg))
	}
	L.SetField(cfg, "argv", argt)
	for k, v := range env.env {
		L.SetField(envt, k, lua.LString(v))
	}
	L.SetField(cfg, "env", envt)
	L.SetField(cfg, "import", L.NewFunction(func(L *lua.LState) int {
		return luaImport(L, env)
	}))
	L.SetGlobal("run", cfg)

	if err := L.DoFile(script); err != nil {
		return fmt.Errorf("failed to run init.lua: %s", err)
	}

	env.argv = []string{}
	argt.ForEach(func(_, v lua.LValue) {
		env.argv = append(env.argv, v.String())
	})
	for k := range env.env {
		delete(env.env, k)
	}
	envt.ForEach(func(k, v lua.LValue) { env.env[k.String()] = v.String() })
	return nil
}

func luaImport(L *lua.LState, env *runEnv) int {
	url := L.CheckString(1)
	if err := importPackage(env, url); err != nil {
		L.RaiseError("failed to import package '%s': %s", url, err)
	}
	return 0
}
