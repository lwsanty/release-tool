# release-tool

This tool is useful if you have a several organizations that contain repositories which constantly required to be released from time to time.

## Collect
This command receives `yml` config file with map where key - owner name, value - array of repositories to discover
```yml
lwsanty:
  - repo0
  - repo1
johnny:
  - repo0
  - repo1
```
Gets releases, post-release commits information and emits it to a file that looks like
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
Where:
- `tag` is the latest release tag name
- `description` is a bullet-list of commits' descriptions that were done after the latest release, if `[no tags]` then all commits' descriptions will be included

Example:
```bash
export GITHUB_TOKEN=XXX
export LOG_LEVEL=debug

go run main.go collect
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
After `collect` is performed you can manually tweak `releases.yml` file to set new iteration of repos versions and correct the descriptions of upcoming releases
When corrections are done you have a new config for a multiple repositories release. Just execute the apply command

Example:
```bash
export GITHUB_TOKEN=XXX
export LOG_LEVEL=debug

go run main.go apply
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