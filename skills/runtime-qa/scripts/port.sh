#!/usr/bin/env bash
# Pick the dev-server port for the current checkout from a port series.
#
#   port.sh <base> [<step>] [<count>]   series: base, base+step, ... (count ports; default step 100, count 10)
#
# Prints one decision line, then one line per port that another checkout holds:
#   reuse <port> slot <n>   a server started from this checkout already listens there
#   start <port> slot <n>   the lowest free port; nothing from this checkout listens in the series
#   taken <port> <dir>      held by another checkout or program (<dir> is its working directory)
#
# Exit: 0 decided, 1 error, 2 every port in the series is taken.

set -uo pipefail

fail() { printf 'error: %s\n' "$*"; exit 1; }
real() { (cd "$1" 2>/dev/null && pwd -P); }

base=${1:-} step=${2:-100} count=${3:-10}
[[ $base =~ ^[0-9]+$ && $step =~ ^[0-9]+$ && $count =~ ^[0-9]+$ ]] || fail "usage: port.sh <base> [<step>] [<count>]"
command -v lsof >/dev/null || fail "lsof not found"
top=$(git rev-parse --show-toplevel 2>/dev/null) || top=$PWD
checkout=$(real "$top")

reuse='' free='' taken=()
for ((i = 0; i < count; i++)); do
  port=$((base + i * step))
  ((port <= 65535)) || break
  pids=$(lsof -nP -t -iTCP:"$port" -sTCP:LISTEN 2>/dev/null | sort -u)
  if [ -z "$pids" ]; then
    [ -n "$free" ] || free="$port slot $i"
    continue
  fi
  owner=''
  for pid in $pids; do
    dir=$(lsof -a -p "$pid" -d cwd -Fn 2>/dev/null | sed -n 's/^n//p' | head -n 1)
    [ -n "$dir" ] || continue
    owner=$dir
    root=$(git -C "$dir" rev-parse --show-toplevel 2>/dev/null) || root=$dir
    if [ "$(real "$root")" = "$checkout" ]; then
      [ -n "$reuse" ] || reuse="$port slot $i"
    fi
  done
  taken+=("$port ${owner:-unknown}")
done

if [ -n "$reuse" ]; then
  printf 'reuse %s\n' "$reuse"
elif [ -n "$free" ]; then
  printf 'start %s\n' "$free"
fi
for t in ${taken[@]+"${taken[@]}"}; do
  case "reuse $t" in "reuse ${reuse%% *} "*) continue ;; esac
  printf 'taken %s\n' "$t"
done
[ -n "$reuse$free" ] || exit 2
