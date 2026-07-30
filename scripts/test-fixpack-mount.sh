#!/usr/bin/env bash
#
# Runs the fixpack mount regression suite (internal/cubecos/fixpacks_mount_test.go).
#
# Those tests mount real filesystem images, so they need root and a free loop
# device. They are behind the `privileged` build tag so `task test` never picks
# them up. This script supplies the environment they need: natively when already
# root on Linux, otherwise inside a privileged container.
#
# Usage: task testFixpackMount   (or ./scripts/test-fixpack-mount.sh)

set -euo pipefail

readonly repoRoot="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly goImage="golang:1.26-alpine"
readonly pkg="./internal/cubecos/"
readonly tags="privileged"

# e2fsprogs-extra carries debugfs, used to build the dirty-journal fixture.
readonly fsTools="e2fsprogs e2fsprogs-extra squashfs-tools util-linux"

runNatively() {
    echo "==> running natively as root on linux"
    cd "$repoRoot"
    CGO_ENABLED=1 GOWORK=off go test -tags="$tags" -count=1 -v "$pkg"
}

runInContainer() {
    if ! command -v docker >/dev/null 2>&1; then
        echo "error: docker is required to run these tests off a Linux root shell" >&2
        exit 1
    fi

    echo "==> running in a privileged $goImage container"
    docker run --rm -i --privileged \
        -v "$repoRoot:/src" \
        -v cube-cos-api-gomodcache:/go/pkg/mod \
        -w /src \
        -e GOWORK=off \
        -e CGO_ENABLED=1 \
        -e GOFLAGS=-buildvcs=false \
        "$goImage" \
        sh -c "apk add --no-cache --quiet $fsTools gcc musl-dev >/dev/null && \
               go test -tags=$tags -count=1 -v $pkg"
}

if [[ "$(uname -s)" == "Linux" && "$(id -u)" -eq 0 ]]; then
    runNatively
else
    runInContainer
fi
