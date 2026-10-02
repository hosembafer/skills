#!/usr/bin/env bash
# Mechanical steps of the ship skill. Run from inside the repo.
#   ship.sh context [target]   facts for the commits and the MR/PR
#   ship.sh push               git push -u origin HEAD, never forced
#   ship.sh open <target>      opens the MR/PR; stdin: title line, blank line, description
# Prints "key: value" lines. Exit: 0 ok, 1 error, 2 usage,
# 3 default or release branch or detached HEAD, 4 behind origin, 5 nothing opened the MR/PR.

export GIT_TERMINAL_PROMPT=0 # fail instead of waiting for a password
release_re='^releases?(/|[-_]v?[0-9]|$)'

top=$(git rev-parse --show-toplevel) || exit 1
cd "$top" || exit 1
branch=$(git symbolic-ref --quiet --short HEAD)
origin_url=$(git remote get-url origin 2>/dev/null)

usage() {
  echo "usage: ship.sh context [target] | push | open <target> < message" >&2
  exit 2
}

on_origin() { git rev-parse -q --verify "refs/remotes/origin/$1" >/dev/null; }

default_branch() {
  local d
  d=$(git symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null) ||
    d=$(git ls-remote --symref origin HEAD 2>/dev/null | awk '$1 == "ref:" { print $2; exit }')
  d=${d#origin/}
  d=${d#refs/heads/}
  [ -n "$d" ] || d=$(git for-each-ref --count=1 --format='%(refname:lstrip=3)' refs/remotes/origin/main refs/remotes/origin/master)
  echo "$d"
}

stop_reason() {
  if [ -z "$branch" ]; then
    echo "detached HEAD"
  elif [ "$branch" = "$default" ]; then
    echo "on the default branch"
  elif printf '%s\n' "$branch" | grep -Eq "$release_re"; then
    echo "on a release branch"
  fi
}

# The default or release branch on origin with the closest merge-base.
infer_target() {
  local c mb d best='' bestd='' tied=''
  for c in $default $(git for-each-ref --format='%(refname:lstrip=3)' refs/remotes/origin | grep -E "$release_re" | grep -vxF "$default"); do
    mb=$(git merge-base HEAD "origin/$c" 2>/dev/null) || continue
    d=$(git rev-list --count "$mb..HEAD")
    if [ -z "$best" ] || [ "$d" -lt "$bestd" ]; then
      best=$c bestd=$d tied=''
    elif [ "$d" -eq "$bestd" ]; then
      tied="$tied $c"
    fi
  done
  if [ -z "$best" ]; then
    echo "unclear (no default or release branch on origin)"
  elif [ -n "$tied" ]; then
    echo "unclear ($best$tied)"
  else
    echo "$best"
  fi
}

# Installed MR/PR CLIs, the one matching the host first.
tools() {
  local t
  case $origin_url in *gitlab*) set -- glab gh ;; *) set -- gh glab ;; esac
  for t; do command -v "$t" >/dev/null && echo "$t"; done
}

find_mr() {
  local t u
  for t in $(tools); do
    case $t in
      gh) u=$(gh pr list --head "$branch" --json url --jq '.[].url' 2>/dev/null | head -1) ;;
      glab) u=$(glab mr list --source-branch "$branch" --output json 2>/dev/null | grep -Eo 'https?://[^"]*/merge_requests/[0-9]+' | head -1) ;;
    esac
    if [ -n "$u" ]; then echo "$u"; return; fi
  done
  if [ -n "$(tools)" ]; then echo none; else echo unknown; fi
}

web_url() {
  printf '%s\n' "$origin_url" | sed -E 's#^(ssh://)?[^@/]+@([^:/]+)(:[0-9]+)?[:/]#https://\2/#; s#^https?://[^@/]+@#https://#; s#\.git$##'
}

templates() {
  git ls-files | grep -iE '^(\.github/|docs/)?pull_request_template\.md$|^\.github/pull_request_template/[^/]+\.md$|^\.gitlab/merge_request_templates/[^/]+\.md$'
}

section() {
  local title=$1 out
  shift
  out=$("$@" 2>&1)
  printf '\n## %s\n%s\n' "$title" "${out:-(none)}"
}

context() {
  local target=$1 stop code=0 ahead behind
  git fetch --quiet origin 2>/dev/null || echo "warning: git fetch origin failed"
  default=$(default_branch)
  if [ -z "$target" ]; then
    target=$(infer_target)
  elif ! on_origin "$target"; then
    target="$target (not on origin)"
  fi
  echo "branch: ${branch:-(detached)}"
  echo "default: ${default:-unknown}"
  echo "target: $target"
  stop=$(stop_reason)
  [ -n "$stop" ] && code=3
  if [ -n "$branch" ] && on_origin "$branch"; then
    read -r ahead behind <<<"$(git rev-list --left-right --count "HEAD...origin/$branch")"
    echo "upstream: origin/$branch (ahead $ahead, behind $behind)"
    if [ "$behind" -gt 0 ] && [ "$code" -eq 0 ]; then
      stop="behind origin/$branch"
      code=4
    fi
    echo "mr: $(find_mr)"
  else
    echo "upstream: none"
    echo "mr: none"
  fi
  [ -n "$stop" ] && echo "stop: $stop"
  section status git status --short --untracked-files=all
  section "diff stat" git diff HEAD --stat
  on_origin "$target" && section "branch commits not on origin/$target" git log --no-merges --format='%h %s' "origin/$target..HEAD"
  section "MR/PR templates" templates
  section "last 30 commits" git log --no-merges -30 --format='%s%n%w(0,4,4)%-b'
  exit "$code"
}

push() {
  local stop out rc ref
  default=$(default_branch)
  stop=$(stop_reason)
  if [ -n "$stop" ]; then echo "stop: $stop"; exit 3; fi
  out=$(git push --porcelain -u origin HEAD 2>&1)
  rc=$?
  ref=$(printf '%s\n' "$out" | awk -F'\t' '$2 ~ /^HEAD:/ { print $1 "|" $3; exit }')
  case $ref in '!|[rejected]'*)
    echo "stop: origin/$branch has commits this branch lacks"
    exit 4
    ;;
  esac
  if [ "$rc" -ne 0 ] || [ -z "$ref" ]; then
    printf '%s\n' "$out" | tail -n 40
    exit 1
  fi
  echo "push: ${ref#*|}"
  # Links the remote prints, for example GitLab's "View merge request" or GitHub's "Create a pull request".
  printf '%s\n' "$out" | tr -d '\r' | sed -n 's/^remote: *//p' | grep -Eo 'https?://[^ ]+' | while read -r u; do
    case $u in
      */merge_requests/[0-9]* | */pull/[0-9]* | */pull-requests/[0-9]*) echo "mr: $u" ;;
      *) echo "create_url: $u" ;;
    esac
  done
}

create_with() {
  case $1 in
    gh) gh pr create --base "$target" --head "$branch" --title="$title" --body="$body" ;;
    glab) glab mr create --source-branch "$branch" --target-branch "$target" --title="$title" --description="$body" --yes ;;
  esac
}

open_mr() {
  local target=$1 title body t out
  [ -n "$target" ] || usage
  IFS= read -r title
  body=$(sed '/./,$!d')
  [ -n "$title" ] || usage
  for t in $(tools); do
    if out=$(create_with "$t" 2>&1); then
      echo "mr: $(printf '%s\n' "$out" | grep -Eo 'https?://[^ ]+' | tail -1)"
      exit 0
    fi
    printf '%s\n' "$out" | sed "s/^/$t: /"
  done
  echo "web: $(web_url)"
  exit 5
}

case ${1-} in
  context) context "${2-}" ;;
  push) push ;;
  open) open_mr "${2-}" ;;
  *) usage ;;
esac
