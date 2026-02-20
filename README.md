# Spark CLI

Workspace CLI for multi-repo development — keeps repositories in sync, manages AWS credentials at the workspace level, and provides dependency-aware builds.

## Install

```bash
brew tap BrianBFarias/spark
brew install spark
```

## Quick Start

```bash
# Create a workspace
spark create workspace ~/MyProject --aws-profile dev --aws-region us-east-1

# Add repos
cd ~/MyProject
spark use my-org/backend-api --build "npm install && npm run build"
spark use my-org/frontend --build "npm install && npm run build" --deps backend-api

# Daily workflow
spark login           # AWS SSO login
spark sync            # git pull all repos
spark build --all     # build in dependency order

# Manage
spark list            # show repos + branch + status
spark workspace       # show workspace info
spark env set KEY=VAL # workspace-wide env vars
```

## Commands

| Command | Description |
|---|---|
| `spark create workspace <path>` | Initialize a new workspace |
| `spark use <org/repo>` | Clone and register a repo |
| `spark sync [repo]` | Pull latest for all or one repo |
| `spark build [repo] --all` | Run build commands (dependency-aware) |
| `spark login` | AWS SSO login using workspace profile |
| `spark list` | List repos with branch and status |
| `spark workspace` | Show workspace details |
| `spark config set --org <name>` | Set global defaults |
| `spark env set/list/export` | Manage workspace env vars |
| `spark remove <repo>` | Unregister a repo from manifest |
| `spark version` | Print version |

## Development

```bash
make build      # build to ./bin/spark
make install    # copy to /usr/local/bin
make test       # run tests
```

## Release

Push a tag to trigger automated release + Homebrew formula update:

```bash
git tag v0.2.0
git push --tags
```

GoReleaser cross-compiles for macOS (Intel + Apple Silicon) and Linux, creates a GitHub release, and auto-updates `Formula/spark.rb` in this repo.
