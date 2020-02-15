#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

IMAGE_NAME=chigusa/bililive-go
VERSION=$(git describe --tags --always)
make PLATFORM=linux ARCH=amd64 bililive
docker build -t $IMAGE_NAME:$VERSION .
docker push $IMAGE_NAME:$VERSION
if ! echo $VERSION | grep "rc" > /dev/null; then
  docker tag $IMAGE_NAME:$VERSION $IMAGE_NAME:latest
  docker push $IMAGE_NAME:latest
fi