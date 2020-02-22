package wiki

import (
	"os"
	"path/filepath"
	"time"

	"github.com/cooper/quiki/wikifier"

	"github.com/cooper/go-git/v4"
	"github.com/cooper/go-git/v4/config"
	"github.com/cooper/go-git/v4/plumbing"
	"github.com/cooper/go-git/v4/plumbing/object"
	"github.com/pkg/errors"
)

// default options for commits made by quiki itself
var quikiCommitOpts = &git.CommitOptions{
	Author: &object.Signature{
		Name:  "quiki",
		Email: "quiki@quiki.app",
		When:  time.Now(),
	},
}

// repo fetches the wiki's git repository, creating it if needed.
func (w *Wiki) repo() (repo *git.Repository, err error) {

	// we've already loaded the repository
	if w._repo != nil {
		repo = w._repo
		return
	}

	// open it
	repo, err = git.PlainOpen(w.Opt.Dir.Wiki)

	// it doesn't exist- let's initialize it
	if err == git.ErrRepositoryNotExists {
		repo, err = w.createRepo()
	} else if err != nil {
		// error in open other than nonexist

		err = errors.Wrap(err, "git:PlainOpen")
		return
	}

	// success
	w._repo = repo
	return
}

// create new repository
func (w *Wiki) createRepo() (repo *git.Repository, err error) {

	/// initialize new repo
	repo, err = git.PlainInit(w.Opt.Dir.Wiki, false)

	// error in init
	if err != nil {
		return nil, errors.Wrap(err, "git:PlainInit")
	}

	// initialized ok

	// create master branch
	err = repo.CreateBranch(&config.Branch{Name: "master"})
	if err != nil {
		return nil, errors.Wrap(err, "git:repo:CreateBranch")
	}

	// TODO: default .gitignore

	// add all files and initial commit
	wt, err := repo.Worktree()
	if err != nil {
		return nil, errors.Wrap(err, "git:repo:Worktree")
	}

	// add .
	_, err = wt.Add(".")
	if err != nil {
		return nil, err
	}

	// commit
	_, err = wt.Commit("Initial commit", quikiCommitOpts)
	if err != nil {
		return nil, errors.Wrap(err, "git:worktree:Commit")
	}

	return repo, nil
}

// // IsRepo returns true if the wiki directory is versioned by git.
// func (w *Wiki) IsRepo() bool {
// 	_, err := os.Stat(filepath.Join(w.Opt.Dir.Wiki, ".git"))
// 	return err == nil
// }

// // InitRepo initializes a git repository for the wiki.
// // If the wiki is already versioned by git, no error is produced.
// func (w *Wiki) InitRepo() error {
// 	if w.IsRepo() {
// 		return nil
// 	}
// 	_, err := git.Init(ginit.Quiet, ginit.Directory(w.Opt.Dir.Wiki))
// 	return err
// }

// BranchNames returns the revision branches available.
func (w *Wiki) BranchNames() ([]string, error) {
	repo, err := w.repo()
	var names []string
	if err != nil {
		return nil, err
	}
	branches, err := repo.Branches()
	if err != nil {
		return nil, err
	}
	branches.ForEach(func(ref *plumbing.Reference) error {
		names = append(names, string(ref.Name()))
		return nil
	})
	return names, nil
}

// ensure a branch exists in git
func (w *Wiki) hasBranch(name string) (bool, error) {
	names, err := w.BranchNames()
	if err != nil {
		return false, err
	}
	for _, branchName := range names {
		if branchName == name {
			return true, nil
		}
	}
	return false, nil
}

// checks out a branch in another directory. returns the directory
func (w *Wiki) checkoutBranch(name string) (string, error) {
	// TODO: make sure name is a simple string with no path elements

	// make cache/branch/ if needed
	wikifier.MakeDir(filepath.Join(w.Opt.Dir.Cache, "branch"), "")

	// e.g. cache/branch/mybranchname
	targetDir := filepath.Join(w.Opt.Dir.Cache, "branch", name)

	// directory already exists, so I'm good with saying the branch is there
	if fi, err := os.Stat(targetDir); err == nil && fi.IsDir() {
		return targetDir, nil
	}

	repo, err := w.repo()
	if err != nil {
		return "", err
	}

	// create the linked repository
	if _, err = repo.PlainAddWorktree(name, targetDir, &git.AddWorktreeOptions{}); err != nil {
		return "", err
	}

	return targetDir, nil
}

// Branch returns a Wiki instance for this wiki at another branch.
// If the branch does not exist, an error is returned.
func (w *Wiki) Branch(name string) (*Wiki, error) {

	// find branch
	if exist, err := w.hasBranch(name); !exist {
		if err != nil {
			return nil, err
		}
		return nil, git.ErrBranchNotFound
	}

	// check out the branch in cache/branch/<name>;
	// if it already has been checked out, this does nothing
	dir, err := w.checkoutBranch(name)
	if err != nil {
		return nil, err
	}

	// create a new Wiki at this location
	// FIXME: this is dumb
	return NewWiki(filepath.Join(dir, "wiki.conf"), "")
}

// NewBranch is like Branch, except it creates the branch at the
// current master revision if it does not yet exist.
func (w *Wiki) NewBranch(name string) (*Wiki, error) {
	repo, err := w.repo()
	if err != nil {
		return nil, err
	}

	// find branch
	if exist, err := w.hasBranch(name); !exist {
		if err != nil {
			return nil, err
		}

		// try to create it
		err := repo.CreateBranch(&config.Branch{
			Name:  name,
			Merge: plumbing.Master,
		})
		if err != nil {
			return nil, err
		}
	}

	// now that it exists, fetch it
	return w.Branch(name)
}
