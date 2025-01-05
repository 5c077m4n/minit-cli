package main

import (
	"minit-cli/builder"
	"minit-cli/cli"
	"minit-cli/store"
)

func must1(err error) {
	if err != nil {
		panic(err)
	}
}

func must2[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func main() {
	cliArgs := must2(cli.GetCLIArgs())

	pkgStore := must2(store.New(cliArgs.Package))
	scriptDir := must2(pkgStore.GetPackageDir(cliArgs.Version))

	must1(builder.Build(cliArgs.Package, scriptDir, builder.BuildTypeFetch))
}
