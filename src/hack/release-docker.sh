#!/bin/sh

set -o errexit
set -o nounset

IMAGE_NAME=chigusa/bililive-go
VERSION=$(git describe --tags --always)

IMAGE_TAG=$IMAGE_NAME:$VERSION

add_latest_tag() {
  if ! echo $VERSION | grep "rc" >/dev/null; then
    echo "-t $IMAGE_NAME:latest"
  fi
}

docker buildx build \
  --platform=linux/amd64,linux/386,linux/arm64/v8,linux/arm/v7,linux/arm/v6 \
  -t $IMAGE_TAG $(add_latest_tag) \
  --build-arg "tag=${VERSION}" \
  --progress plain \
  --push \
  ./
