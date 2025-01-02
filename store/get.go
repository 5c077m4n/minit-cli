package store

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const (
	storeGitName  = "minit-package-store"
	buildFileName = "build.bash"
	fetchFileName = "fetch.bash"
)

type PackageStore struct {
	fs  billy.Filesystem
	dir string
}

var ErrScriptNotFound = errors.New("cloud not find the requested script")

func (ps *PackageStore) getScript(packageName, buildFile string) (string, error) {
	filename := filepath.Join("packages", packageName, buildFile)

	scriptFile, err := ps.fs.Open(filename)
	if err != nil {
		return "", errors.Join(ErrScriptNotFound, err)
	}
	defer scriptFile.Close()

	script, err := io.ReadAll(scriptFile)
	if err != nil {
		return "", errors.Join(ErrScriptNotFound, err)
	}

	return string(script), nil
}

func (ps *PackageStore) GetBuildScript(packageName string) (string, error) {
	return ps.getScript(packageName, buildFileName)
}

func (ps *PackageStore) GetFetchScript(packageName string) (string, error) {
	return ps.getScript(packageName, fetchFileName)
}

func New(commitish string) (*PackageStore, error) {
	tempDir, err := os.MkdirTemp("", "*.minit")
	if err != nil {
		return nil, err
	}

	cloneOpts := &git.CloneOptions{
		URL:               "https://github.com/5c077m4n/" + storeGitName + ".git",
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
	if err := worktree.Checkout(checkoutOpts); err != nil {
		return nil, err
	}

	return &PackageStore{
		fs:  worktree.Filesystem,
		dir: filepath.Join(tempDir, storeGitName),
	}, nil
}
