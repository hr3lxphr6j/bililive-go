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
    res=${file%.exe}.zip
    zip $res ${file} -j ../config.yml >/dev/null 2>&1
    ;;
  tar)
    res=${file}.tar.gz
    tar zcvf $res ${file} -C ../ config.yml >/dev/null 2>&1
    ;;
  7z)
    res=${file}.7z
    7z a $res ${file} ../config.yml >/dev/null 2>&1
    ;;
  *) ;;

  esac
  cd "$last_dir"
  echo $BIN_PATH/$res
}

for dist in $(go tool dist list); do
  case $dist in
  linux/loong64 | android/* | ios/* | js/wasm )
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
  res=$(package $file $package_type)
  rm -f $BIN_PATH/$file
done
