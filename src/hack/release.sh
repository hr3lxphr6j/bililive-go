#!/bin/bash

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


target_platform=("windows" "linux" )
target_arch=("amd64")
# 386,arch,mips,mipsle,riscv64
for dist in $(go tool dist list); do
  echo $dist
  case $dist in
  linux/loong64 | android/* | ios/* | js/wasm )
    continue
    ;;
  *) ;;

  esac
  platform=$(echo ${dist} | cut -d'/' -f1)
  arch=$(echo ${dist} | cut -d'/' -f2)
  echo PLATFORM=${platform} ARCH=${arch}
  # [[ ${target_platform[@]/${platform}/} != ${target_platform[@]} ]]
  # 前一种方式是通过字符串替换后比较数组是否发生变化来判断,而后一种方式是直接使用通配符匹配字符串。
  # 前一种方式更加灵活,可以检查数组中是否存在某个特定值,而后一种方式则更加简单,只能检查数组中是否包含某个子串。
  # 前一种方式对数组的长度和结构没有要求,而后一种方式要求数组元素用空格连接成一个字符串
  if [[ " ${target_platform[*]} " == *${platform}* && " ${target_arch[*]} " == *${arch}* ]]; then   
    echo "build "$dist
    make PLATFORM=${platform} ARCH=${arch} bililive
  fi
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
