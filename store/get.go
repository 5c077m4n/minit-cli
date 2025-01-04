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
	storeGitName = "minit-package-store"
)

type PackageStore struct {
	fs  billy.Filesystem
	dir string
}

var ErrScriptDirNotFound = errors.New("cloud not find the requested package scripts directory")

func (ps *PackageStore) GetPackageDir(packageName string) (string, error) {
	packageDir := filepath.Join("packages", packageName)

	scriptDir, err := ps.fs.Stat(packageDir)
	if err != nil {
		return "", errors.Join(ErrScriptDirNotFound, err)
	}
	if !scriptDir.IsDir() {
		return "", ErrScriptDirNotFound
	}

	return filepath.Join(ps.dir, packageDir), nil
}

func New(commitish string) (*PackageStore, error) {
	stateDir := getStateDir()
	repoDir := filepath.Join(stateDir, "minit", storeGitName)

	cloneOpts := &git.CloneOptions{
		URL:               "https://github.com/5c077m4n/" + storeGitName + ".git",
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress:          os.Stdout,
	}

	repo, err := git.PlainClone(repoDir, false, cloneOpts)

	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		repo, err = git.PlainOpen(repoDir)
		if err != nil {
			return nil, err
		}
		// TODO: add a pull from master here
	}
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

	return &PackageStore{fs: worktree.Filesystem, dir: repoDir}, nil
}
