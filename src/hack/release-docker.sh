#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

make PLATFORM=linux ARCH=amd64 bililive
docker build -t 'chigusa/bililive-go' .