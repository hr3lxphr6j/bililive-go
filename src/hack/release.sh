#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

DISTS=(
    "darwin/386"
    "darwin/amd64"
    "dragonfly/amd64"
    "freebsd/386"
    "freebsd/amd64"
    "freebsd/arm"
    "linux/386"
    "linux/amd64"
    "linux/arm"
    "linux/arm64"
    "linux/ppc64"
    "linux/ppc64le"
    "linux/mips"
    "linux/mipsle"
    "linux/mips64"
    "linux/mips64le"
    "linux/s390x"
    "netbsd/386"
    "netbsd/amd64"
    "netbsd/arm"
    "openbsd/386"
    "openbsd/amd64"
    "openbsd/arm"
    "solaris/amd64"
    "windows/386"
    "windows/amd64"
)
BIN_PATH=bin

package() {
    pushd $BIN_PATH >/dev/null 2>&1
    file=$1
    type=$2
    case $type in
    zip)
        zip ${file%.exe}.zip ${file} ../config.yml
        ;;
    tar)
        tar zcvf ${file}.tar.gz ${file} -C ../ config.yml
        ;;
    7z)
        7z a ${file}.7z ${file} ../config.yml
        ;;
    *) ;;

    esac
    popd >/dev/null 2>&1
}

for dist in ${DISTS[@]}; do
    platform=$(echo ${dist} | cut -d'/' -f1)
    arch=$(echo ${dist} | cut -d'/' -f2)
    make PLATFORM=${platform} ARCH=${arch} bililive
done

for file in $(ls $BIN_PATH); do
    if [[ $file == *".tar.gz" || $file == *".zip" || $file == *".7z" || $file == *".yml" ]]; then
        continue
    fi
    package_type=tar
    if [[ $file == *"windows"* ]]; then
        package_type=zip
    fi
    package $file $package_type
    rm -f $BIN_PATH/$file
done
