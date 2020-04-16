#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
taudir="$workspace/src/github.com/Tau-Coin"
if [ ! -L "$taudir/taucoin-mobile-mining-go" ]; then
    mkdir -p "$taudir"
    cd "$taudir"
    ln -s ../../../../../. taucoin-mobile-mining-go
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$taudir/taucoin-mobile-mining-go"
PWD="$taudir/taucoin-mobile-mining-go"

# Launch the arguments with the configured environment.
exec "$@"
