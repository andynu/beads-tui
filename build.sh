#!/bin/bash
# Build beads-tui binary

set -e

cd "$(dirname "$0")"

go build -o beads-tui ./cmd/beads-tui

echo "Built: ./beads-tui"
