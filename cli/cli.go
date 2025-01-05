package cli

import (
	"errors"
	"flag"
)

type CLIArgs struct {
	Package string
	Version string
}

var ErrNoPackageSelected = errors.New("no package selected")

func GetCLIArgs() (*CLIArgs, error) {
	pkg := flag.String("package", "", "The package you want to install")
	version := flag.String("version", "master", "The branch of the required store")
	flag.Parse()

	if pkg == nil || *pkg == "" {
		return nil, ErrNoPackageSelected
	}

	return &CLIArgs{Package: *pkg, Version: *version}, nil
}
