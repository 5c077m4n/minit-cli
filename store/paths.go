package store

import (
	"os"
)

func getStateDir() string {
	if stateDir, found := os.LookupEnv("XDG_STATE_HOME"); found {
		return stateDir
	} else {
		return os.ExpandEnv("$HOME/.local/state/")
	}
}
