#!/bin/bash
# Install beads-tui to ~/bin

set -e

# Build if binary doesn't exist or is older than source
if [ ! -f ./beads-tui ] || [ ./cmd/beads-tui/main.go -nt ./beads-tui ]; then
    echo "Building beads-tui..."
    go build -o beads-tui ./cmd/beads-tui
fi

# Copy to ~/bin
cp ./beads-tui ~/bin/beads-tui
echo "Installed beads-tui to ~/bin/beads-tui"
