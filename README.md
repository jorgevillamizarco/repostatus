# repostatus

Multi-repo git status viewer. Scans a directory for git repos and shows a color-coded summary table.

## Install

```bash
go install github.com/jorgevillamizar/repostatus@latest
```

Or build from source:

```bash
git clone https://github.com/jorgevillamizar/repostatus
cd repostatus
go build -o repostatus .
```

## Usage

```bash
repostatus                    # scan ~/Documents/projects
repostatus --dir ~/code       # scan a different directory
repostatus --dir /tmp         # any path
```

## Output

Color-coded table showing project name, status, and branch:

```
PROJECT            STATUS         BRANCH
──────────         ──────         ──────
snake              clean          master
rubik              dirty*         feature/ui
portfolio          behind ↓2      main
hermes-agent       no-remote      master
```

## Statuses

| Status | Color | Meaning |
|--------|-------|---------|
| clean | green | Nothing to commit, synced with remote |
| dirty | yellow | Uncommitted changes |
| dirty* | yellow | Uncommitted changes + untracked files |
| ahead ↑N | cyan | N commits ahead of remote |
| behind ↓N | yellow | N commits behind remote |
| diverged ↑N↓M | magenta | Both ahead and behind |
| no-remote | red | No remote configured |
| detached | red | Detached HEAD |
| detached* | red | Detached HEAD + uncommitted changes |
| empty | red | No commits yet |
| error | red | Broken .git directory |

## License

MIT
