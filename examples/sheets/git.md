---
title: Git
description: Everyday Git commands for recording, inspecting, branching, syncing and undoing work.
category: version-control
tags: [git, vcs, branches, history]
source:
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored from the git 2.56.0 documentation; examples are displayed, never executed.
---
Git stores snapshots of a project as a graph of commits. These cards follow a typical day: record changes, look around, move between branches, share work and recover from mistakes.

## Recording changes {#recording}
### Start a repository {#start}
```sh
git init notes
git clone https://example.org/team/notes.git
git clone --depth 1 https://example.org/team/notes.git
```

`--depth 1` fetches only the newest commit, which keeps large histories out of a quick checkout.

### Stage and commit {.col-span-2 #stage-commit}
| Command | Purpose |
| --- | --- |
| `git status --short` | Compact view of changed and untracked paths |
| `git add docs/intro.md` | Stage one file |
| `git add -p` | Choose hunks to stage interactively |
| `git commit -m "docs: explain setup"` | Record the staged snapshot |
| `git commit --amend` | Rewrite the most recent commit with the current index |

### Unstage or discard edits {.warning #restore}
```sh
git restore --staged docs/intro.md   # keep edits, drop from index
git restore docs/intro.md            # throw away working-tree edits
```

The second form overwrites local changes to that file without a backup.

## Inspecting history {.cols-2 #inspecting}
### Read the log {#log}
| Command | Purpose |
| --- | --- |
| `git log --oneline --graph --all` | Branch graph, one line per commit |
| `git log -5 --stat` | Last five commits with changed-file stats |
| `git log --author=Ada --since=2.weeks` | Filter by author and date |
| `git log --grep=fix -i` | Search commit messages, ignoring case |
| `git log -- src/parser.go` | Commits touching one path |

### Compare snapshots {#diff}
| Command | Purpose |
| --- | --- |
| `git diff` | Working tree against the index |
| `git diff --staged` | Index against `HEAD` |
| `git diff main..feature --name-only` | Paths that differ between two branches |
| `git diff --stat HEAD~3` | Summary of changes since three commits ago |

### Who changed a line {#blame}
```sh
git blame -L 40,60 src/parser.go
git show 3f2c1ab
git show HEAD~1:README.md
```

`git show <rev>:<path>` prints a file as it existed in that revision.

## Branches {#branches}
### Create and switch {#switch}
```sh
git switch -c topic/search
git switch main
git switch -          # back to the previous branch
```

### List and tidy {#branch-list}
| Command | Purpose |
| --- | --- |
| `git branch -vv` | Branches with upstream and last commit |
| `git branch --merged main` | Branches already contained in `main` |
| `git branch -d topic/search` | Delete a fully merged branch |
| `git branch -m old-name new-name` | Rename a branch |

`-D` forces deletion even when the branch is unmerged.

### Combine work {.col-span-2 #combine}
```sh
git merge --no-ff topic/search
git rebase main
git rebase -i HEAD~4
git cherry-pick 3f2c1ab
```

Rebasing rewrites commits; avoid it on branches others have already pulled.

## Sharing with remotes {#remotes}
### Remotes and fetching {#fetch}
| Command | Purpose |
| --- | --- |
| `git remote -v` | Show configured remotes and URLs |
| `git remote add upstream <url>` | Track another repository |
| `git fetch --prune` | Update remote-tracking refs, drop deleted ones |
| `git pull --rebase` | Fetch, then replay local commits on top |

### Push {#push}
```sh
git push -u origin topic/search
git push --tags
git push origin --delete topic/search
```

`-u` records the upstream so later `git push` and `git pull` need no arguments.

### Overwrite a remote branch {.warning #force-push}
```sh
git push --force-with-lease origin topic/search
```

Safer than `--force`: it refuses when the remote moved since your last fetch, but it still replaces published history.

## Undo and recover {#undo}
### Reset modes {.warning #reset}
| Command | Effect |
| --- | --- |
| `git reset --soft HEAD~1` | Move the branch back, keep changes staged |
| `git reset HEAD~1` | Move back, keep changes unstaged (`--mixed`) |
| `git reset --hard HEAD~1` | Move back and discard index and working-tree changes |

`--hard` destroys uncommitted work; check `git status` first.

### Revert safely {#revert}
```sh
git revert 3f2c1ab
```

Creates a new commit that undoes an earlier one, so shared history stays intact.

### Shelve work in progress {#stash}
| Command | Purpose |
| --- | --- |
| `git stash push -m "wip: parser"` | Save and clean tracked changes |
| `git stash push -u` | Include untracked files |
| `git stash list` | Show saved entries |
| `git stash apply stash@{1}` | Reapply an entry and keep it |
| `git stash pop` | Reapply the newest entry and drop it |

### Find lost commits {.tip #reflog}
```sh
git reflog
git branch rescue HEAD@{3}
```

The reflog remembers where `HEAD` pointed, even after a reset or rebase.

### Remove untracked files {.warning #clean}
```sh
git clean -n -d    # preview
git clean -f -d    # delete untracked files and directories
```

Add `-x` to also delete ignored files such as build output.
