package main

import (
	"log/slog"
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

	slog.Info(
		"store",
		"value", buildSrc,
	)
}
