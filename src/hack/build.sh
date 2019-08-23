#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

OUTPUT_PATH=bin
SRC_PATH="${GOPATH}/src/github.com/hr3lxphr6j/bililive-go"
CONSTS_PATH="github.com/hr3lxphr6j/bililive-go/src/consts"

_build() {
    target=$1
    bin_name=$2
    ld_flags=$3
    go build \
        -gcflags=${GOFLAGS:-""} \
        -o ${OUTPUT_PATH}/${bin_name} \
        -ldflags="${ld_flags}" \
        ./src/cmd/${target}/
}

build() {
    target=$1

    if [[ ${target} == 'bililive' ]]; then
        now=$(date '+%Y-%m-%d_%H:%M:%S')
        rev=$(echo "${rev:-$(git rev-parse HEAD)}")
        ver=$(git describe --tags --always)
        ld_flags="-s -w -X ${CONSTS_PATH}.BuildTime=${now} -X ${CONSTS_PATH}.AppVersion=${ver} -X ${CONSTS_PATH}.GitHash=${rev}"
    fi

    if [[ $(go env GOOS) == "windows" ]]; then
        ext=".exe"
    fi

    bin_name="${target}-$(go env GOOS)-$(go env GOARCH)${ext:-}"

    _build "${target}" "${bin_name}" "${ld_flags:-}"

    if [[ ${UPX_ENABLE:-"0"} == "1" ]]; then
        upx --no-progress ${OUTPUT_PATH}/"${bin_name}"
    fi
}

main() {
    pushd ${SRC_PATH} >/dev/null
    if [[ ! -d src/cmd/$1 ]]; then
        echo 'Target not exist in src/cmd/'
        popd >/dev/null
        exit 1
    fi
    build $1
    popd >/dev/null
}

main $@
