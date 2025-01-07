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
	must1(err)
	return value
}

func main() {
	cliArgs := must2(cli.GetCLIArgs())

	pkgStore := must2(store.New(cliArgs.Version))
	scriptDir := must2(pkgStore.GetPackageDir(cliArgs.Package))

	must1(builder.BuildShell(cliArgs.Package, scriptDir, builder.BuildTypeFetch))
}
