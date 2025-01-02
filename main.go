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

	fetchScript, err := pkgStore.GetFetchScript("neovim")
	check(err)

	check(builder.Build(fetchScript))
}
