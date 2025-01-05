package store

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const (
	storeRepoName = "minit-package-store"
)

type PackageStore struct {
	fs  billy.Filesystem
	dir string
}

var (
	ErrScriptDirNotFound = errors.New("could not find the requested package scripts directory")
	ErrBadCheckout       = errors.New("could not checkout the requested store version")
	ErrBadPull           = errors.New("could not pull the requested store latest version")
)

func (ps *PackageStore) GetPackageDir(packageName string) (string, error) {
	packageDir := filepath.Join(
		"packages",
		packageName,
		runtime.GOOS,
		runtime.GOARCH,
	)

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
	repoDir := filepath.Join(stateDir, "minit", storeRepoName)

	cloneOpts := &git.CloneOptions{
		URL:               "https://github.com/5c077m4n/" + storeRepoName + ".git",
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress:          os.Stdout,
	}
	repo, err := git.PlainClone(repoDir, false, cloneOpts)

	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		repo, err = git.PlainOpen(repoDir)
	}
	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	checkoutOpts := &git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(commitish),
	}
	if err := worktree.Checkout(checkoutOpts); err != nil {
		return nil, errors.Join(ErrBadCheckout, err)
	}

	err = worktree.Pull(&git.PullOptions{RemoteName: "origin", Force: true})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil, errors.Join(ErrBadPull, err)
	}

	return &PackageStore{fs: worktree.Filesystem, dir: repoDir}, nil
}
