package builder

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

var ErrCouldNotCreatPackageDataDir = errors.New(
	"could not create the data dir for the requested package",
)

func getDataDir() string {
	if stateDir, found := os.LookupEnv("XDG_DATA_HOME"); found {
		return stateDir
	} else {
		return os.ExpandEnv("$HOME/.local/share/")
	}
}

func createPackageBinDir(packageName string) (string, error) {
	dir := filepath.Join(
		getDataDir(),
		"minit",
		"packages",
		packageName,
		runtime.GOOS,
		runtime.GOARCH,
	)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", errors.Join(ErrCouldNotCreatPackageDataDir, err)
	}

	return dir, nil
}
