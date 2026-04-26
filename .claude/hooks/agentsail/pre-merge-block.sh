#!/usr/bin/env sh
set -eu

RUN_JSON="${1:-}"

if [ -z "$RUN_JSON" ] || [ ! -f "$RUN_JSON" ]; then
  echo "Agent Sail: missing run JSON; block merge until agentsail ci has evidence." >&2
  exit 1
fi

if grep -q '"verdict"[[:space:]]*:[[:space:]]*"SHIP"' "$RUN_JSON"; then
  echo "Agent Sail: SHIP verdict found."
  exit 0
fi

echo "Agent Sail: merge blocked; verdict is not SHIP." >&2
exit 1
