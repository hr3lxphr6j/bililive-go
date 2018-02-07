# Bililive-go

直播录制工具，目前支持 哔哩哔哩，熊猫TV，战旗TV

## 依赖
* ffmpeg

## 使用
```
Usage: bililive-go [-hv] [-i urls] [-o path] [-t seconds] [-c filename]
Options:
  -h: this help
  -v: show version and exit
  -i: live room urls, if have many urls, split with "|"
  -o: output file path (default: ./)
  -t: interval of query live status (default: 30)
  -c: set configuration file, command line options with override this (default: ./config.yml)
```