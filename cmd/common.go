package main

import (
	"os"
	"strings"

	"github.com/dorfire/heavenly/pkg/gitutil"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/samber/lo"
	"github.com/urfave/cli/v2"
)

func gitDiff(ctx *cli.Context, pathInRepo string) (gitutil.ChangeSet, error) {
	fromRef := flagOrEnv(ctx, "from-ref", "GITHUB_BASE_REF")
	toCommit := flagOrEnv(ctx, "to-commit", "GITHUB_SHA")

	repo, err := gitutil.OpenRepo(pathInRepo)
	if err != nil {
		return gitutil.ChangeSet{}, err
	}

	repoRoot := lo.Must(repo.Worktree()).Filesystem.Root()
	logger.DebugPrintf("Repo .git path: %s", repoRoot)

	refs, err := repo.References()
	if err != nil {
		return gitutil.ChangeSet{}, err
	}

	err = refs.ForEach(func(r *plumbing.Reference) error {
		if strings.Contains(r.Name().String(), fromRef) {
			logger.DebugPrintf("Possible base ref match: %s", r.Name())
		}
		return nil
	})
	if err != nil {
		return gitutil.ChangeSet{}, err
	}

	return gitutil.FilesChanged(ctx.Context, repo, fromRef, toCommit)
}

func flagOrEnv(ctx *cli.Context, flagName, envVarName string) string {
	res := ctx.String(flagName)
	if res == "" {
		res = os.Getenv(envVarName)
	}
	return res
}
