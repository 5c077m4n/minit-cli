package main

import (
	"minit-cli/builder"
	"minit-cli/store"
)

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func main() {
	pkgStore := must(store.New("master"))
	scriptDir := must(pkgStore.GetPackageDir("neovim"))

	if err := builder.Build(scriptDir, builder.BuildTypeFetch); err != nil {
		panic(err)
	}
}
