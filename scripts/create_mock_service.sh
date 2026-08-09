#!/usr/bin/env bash
set -e

DOMAIN_DIR="${1:-internal/domain}"
PACKAGE_NAME="mocks"

if ! command -v mockgen >/dev/null 2>&1; then
    echo "mockgen not found, installing..."
    go install go.uber.org/mock/mockgen@latest

    GOBIN="$(go env GOPATH)/bin"
    export PATH="$PATH:$GOBIN"

    if ! command -v mockgen >/dev/null 2>&1; then
        echo "mockgen still not found on PATH after install."
        echo "Add this to your shell profile: export PATH=\"\$PATH:$GOBIN\""
        exit 1
    fi
fi

echo "Using: $(command -v mockgen)"

if [ ! -d "$DOMAIN_DIR" ]; then
    echo "Directory $DOMAIN_DIR not found. Run this from the repo root."
    exit 1
fi

for dir in "$DOMAIN_DIR"/*/; do
    dir="${dir%/}"                # strip trailing slash
    repo_file="$dir/repository.go"
    pkg_name="$(basename "$dir")"

    if [ ! -f "$repo_file" ]; then
        echo "skip:  $pkg_name (no repository.go)"
        continue
    fi

    dest_dir="$dir/mocks"
    dest_file="$dest_dir/mock_repository.go"

    mkdir -p "$dest_dir"

    echo "gen:   $repo_file -> $dest_file"
    mockgen -source="$repo_file" \
            -destination="$dest_file" \
            -package="$PACKAGE_NAME"
done

echo "Done."