package gitutil

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type ChangeSet struct {
	Added, Modified, Deleted mapset.Set[string]
}

func OpenRepo(somePathInRepo string) (*git.Repository, error) {
	repo, err := git.PlainOpenWithOptions(somePathInRepo, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, fmt.Errorf("could not detect git repo from path %s: %w", somePathInRepo, err)
	}
	return repo, nil
}

func FilesChanged(ctx context.Context, repo *git.Repository, fromRef string, toCommit string) (ChangeSet, error) {
	baseRef, err := repo.Reference(plumbing.ReferenceName(fromRef), true)
	if err != nil {
		return ChangeSet{}, fmt.Errorf("base reference '%s' not found: %w", fromRef, err)
	}

	baseCommitObj, err := repo.CommitObject(baseRef.Hash())
	if err != nil {
		return ChangeSet{}, fmt.Errorf("base commit '%s' not found: %w", baseRef.Hash(), err)
	}

	toCommitObj, err := repo.CommitObject(plumbing.NewHash(toCommit))
	if err != nil {
		return ChangeSet{}, fmt.Errorf("target commit '%s' not found: %w", toCommit, err)
	}

	diff, err := baseCommitObj.PatchContext(ctx, toCommitObj)
	if err != nil {
		return ChangeSet{}, err
	}

	var pathsAdded, pathsModified, pathsDeleted []string
	for _, f := range diff.FilePatches() {
		f1, f2 := f.Files()
		if f1 == nil {
			pathsAdded = append(pathsAdded, f2.Path())
		} else if f2 == nil {
			pathsDeleted = append(pathsDeleted, f1.Path())
		} else {
			pathsModified = append(pathsModified, f2.Path())
		}
	}

	return ChangeSet{
		mapset.NewSet(pathsAdded...),
		mapset.NewSet(pathsModified...),
		mapset.NewSet(pathsDeleted...),
	}, nil
}

func (s ChangeSet) All() mapset.Set[string] {
	return s.Added.Union(s.Modified).Union(s.Deleted)
}
