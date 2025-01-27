package builder

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type luaFuncPair struct {
	name    string
	luaFunc lua.LGFunction
}

var (
	ErrLuaSetup = errors.New("could not setup the lua build environment")
	ErrLuaBuild = errors.New("could not run the lua build script")
)

func loader(luaState *lua.LState) int {
	exports := map[string]lua.LGFunction{
		"fetchTar": fetchTar,
	}

	mod := luaState.SetFuncs(luaState.NewTable(), exports)
	luaState.SetField(mod, "name", lua.LString("minit"))
	luaState.Push(mod)

	return 1
}

func sanitizeState(luaState *lua.LState) error {
	for _, pair := range []luaFuncPair{
		//{lua.LoadLibName, lua.OpenPackage}, // Must be first(!)
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
	} {
		err := luaState.CallByParam(
			lua.P{
				Fn:      luaState.NewFunction(pair.luaFunc),
				NRet:    0,
				Protect: true,
			},
			lua.LString(pair.name),
		)
		if err != nil {
			return errors.Join(ErrLuaSetup, err)
		}
	}

	return nil
}

func BuildLua(packageName, packagDir string, buildType BuildType) error {
	luaState := lua.NewState()
	defer luaState.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	luaState.SetContext(ctx)
	luaState.PreloadModule("minit", loader)

	if err := sanitizeState(luaState); err != nil {
		return errors.Join(ErrLuaSetup, err)
	}

	luaFilePath := filepath.Join(packagDir, string(buildType)+".lua")
	if err := luaState.DoFile(luaFilePath); err != nil {
		return errors.Join(ErrLuaBuild, err)
	}

	_, err := createPackageBinDir(packageName)
	if err != nil {
		return errors.Join(ErrLuaBuild, err)
	}

	return nil
}
