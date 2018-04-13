#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
lemodir="$workspace/src/github.com/lemochain"
if [ ! -L "$lemodir/lemochain-go" ]; then
    mkdir -p "$lemodir"
    cd "$lemodir"
    ln -s ../../../../../. lemochain-go
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$lemodir/lemochain-go"
PWD="$lemodir/lemochain-go"

# Launch the arguments with the configured environment.
exec "$@"
