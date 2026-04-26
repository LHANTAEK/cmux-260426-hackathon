#!/usr/bin/env bash
set -euo pipefail

CONFIG="${1:-agentsail.loadtest.yaml}"

exec agentsail loadtest run --config "$CONFIG"

