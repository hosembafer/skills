#!/usr/bin/env bash
# Merge the current branch into a shared integration branch through a temporary worktree.
#
#   merge-to.sh [<target>]          push the branch, merge it into <target>, push the merge
#   merge-to.sh --continue <state>  commit the resolved merge and push it
#   merge-to.sh --abort <state>     drop the merge, push nothing
#
# Exit: 0 done, 1 error, 2 stopped, 3 conflicts to resolve, 4 target needed.

set -uo pipefail

say() { printf '%s\n' "$*"; }
stop() { say "stop: $*"; exit 2; }
fail() { say "error: $*"; exit 1; }
wt_git() { git -C "$wt" -c core.hooksPath=/dev/null "$@"; }

repo='' base='' wt='' keep=0 rejected=0
cleanup() {
  [ -n "$wt" ] && [ "$keep" = 0 ] || return 0
  git -C "$repo" worktree remove --force "$wt" 2>/dev/null
  rm -rf "$base"
}
trap cleanup EXIT

save_state() {
  local v
  for v in repo branch target subject src base wt rejected; do printf '%s=%q\n' "$v" "${!v}"; done >"$base/state"
}

default_branch() {
  local ref b
  if ref=$(git symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null); then
    say "${ref#origin/}"
    return
  fi
  for b in main master; do
    git show-ref --verify --quiet "refs/remotes/origin/$b" && { say "$b"; return; }
  done
}

is_protected() {
  case $1 in "$default" | release/* | release-*) return 0 ;; esac
  return 1
}

fetch_target() {
  local err
  err=$(git fetch --quiet origin "+refs/heads/$target:refs/remotes/origin/$target" 2>&1) && return 0
  case $err in *"couldn't find remote ref"*) stop "origin/$target doesn't exist" ;; esac
  fail "fetching $target failed: $err"
}

# Integration branches this branch was merged into before, newest first.
merge_targets() {
  local re="^Merge (remote-tracking )?branch '(origin/)?([^']+)'( of [^ ]+)? into '?([^']+)'?\$" s
  git log --remotes --merges -F --grep="'$branch'" --format=%s | while IFS= read -r s; do
    [[ $s =~ $re ]] || continue
    [ "${BASH_REMATCH[3]}" = "$branch" ] || continue
    is_protected "${BASH_REMATCH[5]}" || say "${BASH_REMATCH[5]}"
  done | awk '!seen[$0]++'
}

report_conflicts() {
  local f
  keep=1
  save_state
  say "conflicts: $base"
  say "worktree: $wt"
  wt_git diff --name-only --diff-filter=U | while IFS= read -r f; do
    say "file: $f"
    wt_git log --merge --left-right -n 6 --format='  %m %h %s' -- "$f"
  done
  exit 3
}

merge() {
  local out
  if ! out=$(wt_git merge --no-ff --no-edit -m "$subject" "$src" 2>&1); then
    [ -n "$(wt_git diff --name-only --diff-filter=U)" ] || fail "merge failed: $out"
    report_conflicts
  fi
  case $out in *"lready up"*"to"*"date"*)
    say "merge: nothing to merge, origin/$target already has $branch"
    exit 0
    ;;
  esac
  push_target
}

push_target() {
  local err
  if err=$(wt_git push --quiet origin "HEAD:refs/heads/$target" 2>&1); then
    say "merge: $(wt_git log -1 --format='%h %s')"
    say "target: pushed to origin/$target"
    exit 0
  fi
  case $err in *rejected* | *"fetch first"* | *non-fast-forward*) ;; *) fail "pushing $target failed: $err" ;; esac
  rejected=$((rejected + 1))
  [ "$rejected" -ge 2 ] && stop "origin/$target moved twice during the merge; nothing was pushed to it"
  say "target: origin/$target moved, merging again"
  fetch_target
  wt_git checkout --quiet --detach "origin/$target" || fail "resetting the worktree failed"
  merge
}

start() {
  local target_arg=${1:-} merged_into n dirty last
  branch=$(git symbolic-ref --quiet --short HEAD) || stop "HEAD is detached"
  repo=$(git rev-parse --show-toplevel) || fail "not inside a git repository"
  default=$(default_branch)
  target=$target_arg
  if [ -z "$target" ]; then
    merged_into=$(merge_targets)
    n=$(printf '%s' "$merged_into" | grep -c .)
    if [ "$n" -eq 1 ]; then
      target=$merged_into
      say "target: $target (where earlier merges of $branch went)"
    else
      say "target: needed"
      exit 4
    fi
  fi
  [ "$target" = "$branch" ] && stop "$target is the current branch"
  is_protected "$target" && stop "$target is the default or a release branch; those go through an MR/PR"
  fetch_target

  # The merge takes origin/<branch>; set an upstream only when the branch has none.
  up=origin/$branch
  push=(git push --quiet origin HEAD)
  git rev-parse --quiet --verify '@{u}' >/dev/null || push=(git push --quiet -u origin HEAD)
  if git fetch --quiet origin "+refs/heads/$branch:refs/remotes/$up" 2>/dev/null; then
    read -r behind ahead <<<"$(git rev-list --left-right --count "$up...HEAD")"
    [ "$behind" -gt 0 ] && [ "$ahead" -gt 0 ] && stop "$branch has diverged from $up"
    [ "$behind" -gt 0 ] && stop "$branch is behind $up"
    if [ "$ahead" -gt 0 ]; then
      "${push[@]}" || fail "pushing $branch failed"
      say "branch: pushed $ahead commit(s) to $up"
    else
      say "branch: $up is up to date"
    fi
  else
    "${push[@]}" || fail "pushing $branch failed"
    say "branch: pushed to $up"
  fi
  src=$(git rev-parse --verify --quiet "$up") || fail "can't resolve $up"

  dirty=$(git status --porcelain)
  if [ -n "$dirty" ]; then
    say "uncommitted: left out of the merge"
    printf '%s\n' "$dirty" | sed 's/^/  /'
  fi

  subject="Merge branch '$branch' into $target"
  last=$(git log "origin/$target" --merges -20 --format=%s | grep -m1 -E "^Merge branch '.+' into ")
  case $last in *"into '"*"'") subject="Merge branch '$branch' into '$target'" ;; esac

  base=$(mktemp -d "${TMPDIR:-/tmp}/merge-to.XXXXXX") || fail "can't create a temporary directory"
  git -c core.hooksPath=/dev/null worktree add --quiet --detach "$base/tree" "origin/$target" || fail "creating the worktree failed"
  wt=$base/tree
  merge
}

resume() {
  [ -f "${2:-}/state" ] || fail "no merge state in ${2:-<missing>}"
  # shellcheck disable=SC1091
  . "$2/state"
  cd "$repo" || fail "can't enter $repo"
  if [ "$1" = --abort ]; then
    wt_git merge --abort 2>/dev/null
    say "merge: aborted, nothing pushed to origin/$target"
    exit 0
  fi
  wt_git add -A
  if [ -n "$(wt_git diff --cached --check | grep 'conflict marker')" ] || [ -n "$(wt_git diff --name-only --diff-filter=U)" ]; then
    say "unresolved: conflict markers are left"
    wt_git diff --cached --check | grep 'conflict marker' | sed 's/^/  /'
    keep=1
    exit 3
  fi
  wt_git commit --quiet -m "$subject" || fail "committing the merge failed"
  push_target
}

case ${1:-} in
--continue | --abort) resume "$@" ;;
*) start "${1:-}" ;;
esac
