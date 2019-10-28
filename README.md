# release-tool

This tool provides an easy way to batch releases on multiple Github repositories across organizations.

It works in three stages:
- Collect commit logs and latest release tags (automated)
- Edit the release file to set description and the new version tag (manual)
- Apply (push) the releases (automated)

## Collect

First, create a file with a list of repositories:

```yml
lwsanty:
  - repo0
  - repo1
johnny:
  - repo0
  - repo1
```

To collect an information for upcoming releases, run the following command: 

```bash
export GITHUB_TOKEN=XXX
relese-tool collect -c config.yml
```

It will generate a `releases.yml` file that contains commit logs for all repositories, as well as latest release tags (if any):

```yml
lwsanty:
  repo0:
    tag: 'v0.0.1'
    description: |-
      * post-release comment 0
      * post-release comment 1
  repo1:
    tag: '[no tags]'
    description: |-
      * post-release comment 0
      * post-release comment 1
johnny:
  repo0:
    tag: 'v0.0.1'
    description: |-
      * post-release comment 0
      * post-release comment 1
  repo1:
    tag: '[no tags]'
    description: |-
      * post-release comment 0
      * post-release comment 1
```

The `tag` is the latest existing release tag. You must increment a specific version number there, since the tool cannot be sure if the change minor or not (e.g. minor).

The `description` is a release description. It is generated as a bullet list of commit messages that were done after the latest release.
If there was no tags (as indicated by `tag: '[no tags]'`), then all commits will be included.
This serves as a baseline, you are free to rewrite the description according to a preferred style.

A full example:
```bash
export GITHUB_TOKEN=XXX
export LOG_LEVEL=debug

release-tool collect
```
Usage:
```
Usage:
  release-tool [OPTIONS] collect [collect-OPTIONS]

Help Options:
  -h, --help         Show this help message

[collect command options]
          --dry-run  performs extra debug info instead of the real action
      -c, --config=  path to file with owners and repositories (default: config.yml) [$CONFIG]
      -f, --file=    path to emit file with releases info (default: releases.yml) [$FILE]
```

## Apply

After editing the `releses.yml` you can send (apply) those releases to Github. Just execute the `apply` command:

Example:
```bash
export GITHUB_TOKEN=XXX
export LOG_LEVEL=debug

relese-tool apply
```

Usage:
```
Usage:
  release-tool [OPTIONS] apply [apply-OPTIONS]

Help Options:
  -h, --help                Show this help message

[apply command options]
          --dry-run         performs extra debug info instead of the real action
      -f, --file=           path to file with configuration (default: releases.yml) [$FILE]
          --release-branch= branch to release (default: master) [$RELEASE_BRANCH]
```