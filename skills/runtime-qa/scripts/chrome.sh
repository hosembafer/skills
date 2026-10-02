#!/usr/bin/env bash
# Open a URL in the isolated Chrome profile of a QA slot (macOS).
#
#   chrome.sh [--agent claude|codex] <slot> [<url>]   start the slot's Chrome, or open <url> in it when it already runs
#   chrome.sh --list                                  every slot profile: running or idle, and each agent's extension
#
# Each slot is a separate Chrome instance with its own --user-data-dir under ~/.runtime-qa/chrome,
# so cookies, storage and sign-ins never leak between slots or into the user's own Chrome.
# --agent picks the extension the session drives Chrome through: Claude in Chrome (default) or Codex's ChatGPT extension.
#
# Prints:
#   profile <dir>
#   chrome launched|running
#   extension installed|missing   missing: the agent's extension store page was opened too
#   id <id>|none                  the id the agent selects this browser by; none until the extension has connected
#
# --list prints one line per slot: slot-<n> running|idle claude=<id|none|missing> codex=<id|none|missing>
#
# Exit: 0 done, 1 error.

set -uo pipefail

fail() { printf 'error: %s\n' "$*"; exit 1; }
usage() { fail "usage: chrome.sh [--agent claude|codex] <slot> [<url>] | chrome.sh --list"; }

root="$HOME/.runtime-qa/chrome"
app='Google Chrome'
hosts="$HOME/Library/Application Support/Google/Chrome/NativeMessagingHosts"

# use_agent <agent>: its extension id, store page, and the chrome.storage key and pattern of the id it selects Chrome by.
use_agent() {
  case "$1" in
    claude)
      ext='fcoeoabgfenejglbffodgkkbkcdhcgfn' store="https://chromewebstore.google.com/detail/$ext"
      key='bridgeDeviceId' pattern='[0-9a-f-]{36}'
      ;;
    codex)
      ext='hehggadaopoacecdllhhajmbjkdcmajg' store="https://chromewebstore.google.com/detail/chatgpt/$ext"
      key='extensionInstanceId' pattern='[A-Za-z0-9_-]{1,128}'
      ;;
    *) fail "unknown agent: $1 (claude or codex)" ;;
  esac
}

running() { pgrep -f -- "--user-data-dir=$1( |\$)" >/dev/null 2>&1; }
has_ext() { compgen -G "$1/*/Extensions/$ext" >/dev/null; }

# agent_id <dir>: the newest stored id, none before the extension has connected, missing when it isn't installed.
agent_id() {
  local id
  has_ext "$1" || { printf missing; return; }
  id=$(LC_ALL=C grep -aohE "$key.{0,4}\"$pattern\"" "$1"/*/"Local Extension Settings/$ext"/* 2>/dev/null |
    tail -n 1 | sed -E "s/^$key.{0,4}\"//; s/\"\$//")
  printf '%s' "${id:-none}"
}

if [ "${1:-}" = --list ]; then
  for dir in "$root"/slot-*; do
    [ -d "$dir" ] || continue
    state=idle
    running "$dir" && state=running
    use_agent claude && claude=$(agent_id "$dir")
    use_agent codex && codex=$(agent_id "$dir")
    printf '%s %s claude=%s codex=%s\n' "${dir##*/}" "$state" "$claude" "$codex"
  done
  exit 0
fi

agent=claude
if [ "${1:-}" = --agent ]; then
  [ $# -ge 2 ] || usage
  agent=$2
  shift 2
fi
use_agent "$agent"

slot=${1:-} url=${2:-}
[[ $slot =~ ^[0-9]+$ ]] || usage
[ "$(uname -s)" = Darwin ] || fail "macOS only"
open -Ra "$app" 2>/dev/null || fail "$app is not installed"

dir="$root/slot-$slot"
mkdir -p "$dir/NativeMessagingHosts" || fail "cannot create $dir"

# Chrome reads native messaging hosts from its user data dir, so a separate profile needs the agents' hosts linked in.
for host in "$hosts"/com.anthropic.*.json "$hosts"/com.openai.codexextension.json; do
  [ -e "$host" ] && ln -sf "$host" "$dir/NativeMessagingHosts/"
done

urls=()
[ -n "$url" ] && urls+=("$url")
installed=installed
has_ext "$dir" || { installed=missing; urls+=("$store"); }

state=launched
running "$dir" && state=running
open -na "$app" --args --user-data-dir="$dir" --no-first-run --no-default-browser-check ${urls[@]+"${urls[@]}"} ||
  fail "could not start $app"

id=$(agent_id "$dir")
[ "$id" = missing ] && id=none
printf 'profile %s\nchrome %s\nextension %s\nid %s\n' "$dir" "$state" "$installed" "$id"
