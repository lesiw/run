package main

import (
	"fmt"
	"os"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func runInit() error {
	if _, err := os.Stat(".run/init.lua"); err != nil {
		return nil
	}

	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer L.Close()

	cfg := L.NewTable()
	initGetEnv(L, cfg)
	L.SetGlobal("run", cfg)

	if err := L.DoFile(".run/init.lua"); err != nil {
		return fmt.Errorf("failed to run init.lua: %s", err)
	}

	if err := initSetEnv(L, cfg); err != nil {
		return fmt.Errorf("failed to apply env from init.lua: %s", err)
	}

	return nil
}

func initGetEnv(L *lua.LState, t *lua.LTable) {
	envtbl := L.NewTable()
	for _, env := range os.Environ() {
		parts := strings.Split(env, "=")
		key := parts[0]
		value := parts[1]
		L.SetField(envtbl, key, lua.LString(value))
	}
	L.SetField(t, "env", envtbl)
}

func initSetEnv(L *lua.LState, t *lua.LTable) error {
	v := L.GetField(t, "env")
	envtbl, ok := v.(*lua.LTable)
	if !ok {
		return fmt.Errorf("run.env is not a table")
	}
	envtbl.ForEach(func(k lua.LValue, v lua.LValue) {
		os.Setenv(k.String(), v.String())
	})
	return nil
}
