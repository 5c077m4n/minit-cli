package store

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const (
	buildFileName = "build.bash"
	fetchFileName = "fetch.bash"
)

type PackageStore struct {
	fs billy.Filesystem
}

var ErrScriptNotFound = errors.New("cloud not file the requested script")

func (ps *PackageStore) getScript(packageName, buildType string) (string, error) {
	filename := filepath.Join("packages", packageName, buildType)

	scriptFile, err := ps.fs.Open(filename)
	if err != nil {
		return "", errors.Join(ErrScriptNotFound, err)
	}
	defer scriptFile.Close()

	content := make([]byte, 4096)
	readLength, err := scriptFile.Read(content)
	if err != nil {
		return "", errors.Join(ErrScriptNotFound, err)
	}

	return string(content[:readLength]), nil
}

func (ps *PackageStore) GetBuildScript(packageName string) (string, error) {
	return ps.getScript(packageName, buildFileName)
}

func (ps *PackageStore) GetFetchScript(packageName string) (string, error) {
	return ps.getScript(packageName, fetchFileName)
}

func New(commitish string) (*PackageStore, error) {
	tempDir, err := os.MkdirTemp("", "minit")
	if err != nil {
		return nil, err
	}

	cloneOpts := &git.CloneOptions{
		URL:               "https://github.com/5c077m4n/minit-package-store.git",
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress:          os.Stdout,
	}

	repo, err := git.PlainClone(tempDir, false, cloneOpts)
	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	checkoutOpts := &git.CheckoutOptions{Hash: plumbing.NewHash(commitish)}
	err = worktree.Checkout(checkoutOpts)
	if err != nil {
		return nil, err
	}

	return &PackageStore{fs: worktree.Filesystem}, nil
}
