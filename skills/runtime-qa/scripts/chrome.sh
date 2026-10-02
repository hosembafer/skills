#!/usr/bin/env bash
# Open a URL in the isolated Chrome profile of a QA slot (macOS).
#
#   chrome.sh <slot> [<url>]   start the slot's Chrome, or open <url> in it when it already runs
#   chrome.sh --list           every slot profile: running or idle, Claude installed or not, its device id
#
# Each slot is a separate Chrome instance with its own --user-data-dir under ~/.runtime-qa/chrome,
# so cookies, storage and sign-ins never leak between slots or into the user's own Chrome.
#
# Prints:
#   profile <dir>
#   chrome launched|running
#   extension installed|missing   missing: the Claude extension's store page was opened too
#   device <id>|none              the deviceId this profile's Claude extension connects as; none before its first sign-in
#
# Exit: 0 done, 1 error.

set -uo pipefail

fail() { printf 'error: %s\n' "$*"; exit 1; }

root="$HOME/.runtime-qa/chrome"
app='Google Chrome'
ext_id='fcoeoabgfenejglbffodgkkbkcdhcgfn'
store="https://chromewebstore.google.com/detail/$ext_id"

running() { pgrep -f -- "--user-data-dir=$1( |\$)" >/dev/null 2>&1; }
has_ext() { compgen -G "$1/*/Extensions/$ext_id" >/dev/null; }
# The extension keeps its deviceId under bridgeDeviceId in its chrome.storage files; the newest file's value wins.
device() {
  local id
  id=$(LC_ALL=C grep -aohE 'bridgeDeviceId.{0,4}"[0-9a-f-]{36}"' "$1"/*/"Local Extension Settings/$ext_id"/* 2>/dev/null |
    tail -n 1 | grep -oE '[0-9a-f-]{36}')
  printf '%s' "${id:-none}"
}

if [ "${1:-}" = --list ]; then
  for dir in "$root"/slot-*; do
    [ -d "$dir" ] || continue
    state=idle ext=missing
    running "$dir" && state=running
    has_ext "$dir" && ext=installed
    printf '%s %s extension %s device %s\n' "${dir##*/}" "$state" "$ext" "$(device "$dir")"
  done
  exit 0
fi

slot=${1:-} url=${2:-}
[[ $slot =~ ^[0-9]+$ ]] || fail "usage: chrome.sh <slot> [<url>] | chrome.sh --list"
[ "$(uname -s)" = Darwin ] || fail "macOS only"
open -Ra "$app" 2>/dev/null || fail "$app is not installed"

dir="$root/slot-$slot"
mkdir -p "$dir/NativeMessagingHosts" || fail "cannot create $dir"

# Chrome reads native messaging hosts from its user data dir, so a separate profile needs Claude's host linked in.
for host in "$HOME/Library/Application Support/Google/Chrome/NativeMessagingHosts"/com.anthropic.*.json; do
  [ -e "$host" ] && ln -sf "$host" "$dir/NativeMessagingHosts/"
done

urls=()
[ -n "$url" ] && urls+=("$url")
ext=installed
has_ext "$dir" || { ext=missing; urls+=("$store"); }

state=launched
running "$dir" && state=running
open -na "$app" --args --user-data-dir="$dir" --no-first-run --no-default-browser-check ${urls[@]+"${urls[@]}"} ||
  fail "could not start $app"

printf 'profile %s\nchrome %s\nextension %s\ndevice %s\n' "$dir" "$state" "$ext" "$(device "$dir")"
