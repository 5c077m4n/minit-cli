package main

import (
	"minit-cli/builder"
	"minit-cli/store"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	pkgStore, err := store.New("master")
	check(err)

	scriptDir, err := pkgStore.GetPackageDir("neovim")
	check(err)

	check(builder.Build(scriptDir, builder.BuildTypeFetch))
}
