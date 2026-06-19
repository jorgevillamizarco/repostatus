# repostatus

At-a-glance git status for all your repositories. Scans a directory tree, finds every git repo, and shows a color-coded status table.

## Install

Requires Go 1.21+.

```bash
cd ~/Documents/projects/repostatus
go build .
```

Optionally, symlink to your PATH:

```bash
ln -s ~/Documents/projects/repostatus/repostatus ~/.local/bin/repostatus
```

## Usage

```bash
# Scan default directory (~/Documents/projects)
./repostatus

# Scan a specific directory
./repostatus --dir /path/to/projects
```

## Output

```
PROJECT               STATUS           BRANCH
───────────────────── ──────────────── ──────────────────────────
hermes-agent-team     clean            master
deep-research         dirty*           feature/structured-reports
portfolio-agent       clean ahead ↑2   main
rubik                 no-remote        master
```

Columns: **PROJECT** | **STATUS** | **BRANCH**

## Statuses

| Status | Meaning |
|--------|---------|
| `clean` | No uncommitted changes, synced with remote |
| `clean ahead ↑N` | N unpushed commits ahead of remote |
| `clean behind ↓N` | N unpulled commits behind remote |
| `dirty` | Uncommitted changes (modified or staged files) |
| `dirty*` | Uncommitted changes + untracked files |
| `no-remote` | No remote configured (origin not found) |
| `detached` | Detached HEAD (on a specific commit, no branch) |
| `detached*` | Detached HEAD with uncommitted changes |
| `empty` | Repository has no commits |
| `error` | `.git` directory exists but `git status` failed |

Ahead/behind/dirty can combine: `dirty ahead ↑1`, `clean behind ↓2`.

## Color Scheme

- **Green**: clean, synced
- **Yellow**: dirty (local changes), behind remote
- **Cyan**: ahead of remote
- **Magenta**: diverged (both ahead and behind)
- **Red**: no-remote, detached, empty, error

## Exit Codes

- `0` — successful scan (even if no repos found)
- `1` — error (bad directory path, permission denied)

## Edge Cases Handled

- Empty repos (no commits)
- No remote configured
- Detached HEAD (commit hash shown, branch is the short SHA)
- Detached HEAD with dirty files (marked `detached*`)
- Bare repos (skipped, shown as `bare`)
- Spaces in directory names
- Nested git repos (each shown independently)
- Very long project names (auto-adjusted column widths)
