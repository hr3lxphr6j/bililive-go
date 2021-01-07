#!/bin/sh

set -o errexit
set -o nounset

readonly BIN_PATH=bin

package() {
  last_dir=$(pwd)
  cd $BIN_PATH
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
  cd "$last_dir"
}

for dist in $(go tool dist list); do
  case $dist in
  android/* | darwin/arm64 | js/wasm)
    continue
    ;;
  *) ;;

  esac
  platform=$(echo ${dist} | cut -d'/' -f1)
  arch=$(echo ${dist} | cut -d'/' -f2)
  make PLATFORM=${platform} ARCH=${arch} bililive
done

for file in $(ls $BIN_PATH); do
  case $file in
  *.tar.gz | *.zip | *.7z | *.yml | *.yaml)
    continue
    ;;
  *windows*)
    package_type=zip
    ;;
  *)
    package_type=tar
    ;;
  esac
  package $file $package_type
  rm -f $BIN_PATH/$file
done
