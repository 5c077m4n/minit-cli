package main

import (
	"minit-cli/builder"
	"minit-cli/store"
)

func main() {
	pkgStore, err := store.New("master")
	if err != nil {
		panic(err)
	}

	buildSrc, err := pkgStore.GetFetchScript("neovim")
	if err != nil {
		panic(err)
	}

	err = builder.Build(buildSrc)
	if err != nil {
		panic(err)
	}
}
