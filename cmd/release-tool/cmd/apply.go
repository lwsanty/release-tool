package cmd

import (
	"context"
	"io/ioutil"

	"github.com/google/go-github/v28/github"
	"gopkg.in/src-d/go-errors.v1"
	"gopkg.in/src-d/go-log.v1"
	"gopkg.in/yaml.v2"
)

const ApplyCommandDescription = "takes YAML file, generated with collect command and edited with new releases info and performs new releases if it's required"

var errFailedToUpdateAllDrivers = errors.NewKind("release failed for one or several drivers, last error: %v")

type ApplyCommand struct {
	DryRun        bool   `long:"dry-run" description:"performs extra debug info instead of the real action"`
	File          string `long:"file" short:"f" env:"FILE" default:"releases.yml" description:"path to file with configuration"`
	ReleaseBranch string `long:"release-branch" env:"RELEASE_BRANCH" default:"master" description:"branch to release"`
}

func (c *ApplyCommand) Execute(args []string) error {
	ctx := context.Background()

	data, err := ioutil.ReadFile(c.File)
	if err != nil {
		return err
	}

	owners := make(Owners)
	if err := yaml.Unmarshal(data, owners); err != nil {
		return err
	}

	var lastErr error
	for owner, repos := range owners {
		for repo, release := range repos {
			if err := c.release(ctx, owner, repo, release); err != nil {
				lastErr = err
				log.Errorf(err, "failed to release %s/%s => %s", owner, repo, release.Tag)
				continue
			}
		}
	}
	if lastErr != nil {
		return errFailedToUpdateAllDrivers.New(lastErr)
	}

	return nil
}

// TODO: default branch?
// release uses github API to create the release from previously parsed config file
func (c *ApplyCommand) release(ctx context.Context, owner, repo string, release Release) error {
	client := githubClient()

	releaseConfig := &github.RepositoryRelease{
		TagName:         github.String(release.Tag),
		TargetCommitish: github.String(c.ReleaseBranch),
		Name:            github.String(release.Tag),
		Body:            github.String(release.Description),
		Draft:           github.Bool(false),
		Prerelease:      github.Bool(false),
	}

	if c.DryRun {
		data, err := yaml.Marshal(releaseConfig)
		if err != nil {
			return err
		}
		log.Infof("%s: release config:\n%v\n", repo, string(data))
		return nil
	}

	releaseResp, _, err := client.Repositories.CreateRelease(ctx, owner, repo, releaseConfig)
	if err != nil {
		return err
	}
	log.Infof("%v: release %v has been successfully created", repo, *releaseResp.Name)
	return nil
}
