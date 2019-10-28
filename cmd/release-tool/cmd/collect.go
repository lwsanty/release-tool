package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-errors.v1"
	"gopkg.in/src-d/go-log.v1"
	"gopkg.in/yaml.v2"
)

const (
	CollectCommandDescription = "pulls info about all latest releases of repositories and all commits since those releases and dumps this info to a YAML file"

	noTags = "[no tags]"
)

type CollectCommand struct {
	DryRun bool   `long:"dry-run" description:"performs extra debug info instead of the real action"`
	Config string `long:"config" short:"c" env:"CONFIG" default:"config.yml" description:"path to file with owners and repositories"`
	File   string `long:"file" short:"f" env:"FILE" default:"releases.yml" description:"path to emit file with releases info"`
}

// Owners represents map with
// - key: owner name
// - value: key
type Owners map[string]OwnerReleases

// key represents map with
// - key: repository name
// - value: Release object
type OwnerReleases map[string]Release

// Release represents an object with the latest tag + concatenated descriptions of commits, performed after release
type Release struct {
	// Tag corresponds to release tag in format v0.0.1
	Tag string `yaml:"tag"`
	// Description represents concatenated descriptions of commits
	Description string `yaml:"description"`
}

var errFailedToGetCommitsDescription = errors.NewKind("%v: failed to get commits description: %v")

var gh struct {
	once   sync.Once
	client *github.Client
}

func (c *CollectCommand) Execute(args []string) error {
	repositories := make(map[string][]string)
	data, err := ioutil.ReadFile(c.Config)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, repositories); err != nil {
		return err
	}

	ctx := context.Background()
	result := make(Owners)
	for owner, repos := range repositories {
		for _, r := range repos {
			tag, err := getRepoReleaseTag(ctx, owner, r)
			if err != nil {
				return err
			}
			desc, err := getCommitsDescription(ctx, owner, r, tag)
			if err != nil {
				return err
			}

			if _, ok := result[owner]; !ok {
				result[owner] = make(OwnerReleases)
			}

			result[owner][r] = Release{
				Tag:         tag,
				Description: desc,
			}
		}
	}

	resData, err := yaml.Marshal(result)
	if err != nil {
		return err
	}

	if c.DryRun {
		log.Infof("\n%v\n", string(resData))
		return nil
	}
	if err := ioutil.WriteFile(c.File, resData, 0644); err != nil {
		return err
	}

	log.Infof("file %v has been successfully written", c.File)
	return nil
}

// getRepoReleaseTag returns tag name of the latest release, if no releases contained - return noTags
func getRepoReleaseTag(ctx context.Context, owner, repo string) (string, error) {
	client := githubClient()
	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return noTags, nil
		}
		return "", err
	}

	return release.GetTagName(), nil
}

// getCommitsDescription is used to obtain concatenated descriptions of commits, performed after release
func getCommitsDescription(ctx context.Context, owner, repo, tag string) (string, error) {
	client := githubClient()

	var listOpts github.CommitsListOptions
	if tag != noTags {
		release, _, err := client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
		if err != nil {
			return "", errFailedToGetCommitsDescription.New(repo, err)
		}
		listOpts.Since = release.PublishedAt.UTC()
	}

	var results []string
	for page := 1; ; page++ {
		opt := listOpts
		opt.ListOptions = github.ListOptions{
			Page: page, PerPage: 100,
		}

		rCommits, _, err := client.Repositories.ListCommits(ctx, owner, repo, &opt)
		if err != nil {
			return "", errFailedToGetCommitsDescription.New(repo, err)
		}

		log.Debugf("commits: %v", len(rCommits))
		if len(rCommits) == 0 {
			break
		}
		for _, c := range rCommits {
			sha := c.GetSHA()
			belongs, err := checkCommit(ctx, owner, repo, sha)
			if err != nil {
				log.Errorf(err, "failed to check commit")
				continue
			}
			if !belongs {
				log.Debugf("%s/%s: commit %s does not belong to default branch", owner, repo, sha)
				continue
			}

			log.Debugf("%s/%s: appending: %v", owner, repo, *c.Commit.Message)
			results = append(results, processCommitMessage(*c.Commit.Message))
		}
	}

	return strings.Join(results, "\n"), nil
}

// checkCommit checks if commit belongs to default branch
func checkCommit(ctx context.Context, owner, repo, hash string) (bool, error) {
	client := githubClient()

	query := fmt.Sprintf("repo:%s/%s hash:%s", owner, repo, hash)
	res, _, err := client.Search.Commits(ctx, query, nil)
	if err != nil {
		return false, err
	}

	return res.GetTotal() > 0, nil
}

// githubClient returns a lazily-initialized singleton Github API client.
func githubClient() *github.Client {
	gh.once.Do(func() {
		hc := http.DefaultClient
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			hc = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			))
		}
		gh.client = github.NewClient(hc)
	})
	return gh.client
}

// processCommitMessage performs required formatting to represent concatenated commits as a bullet list
func processCommitMessage(msg string) string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(msg))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "Signed-off") || line == "" {
			continue
		}
		lines = append(lines, line)
	}

	return "* " + strings.Join(lines, "\n")
}
