# Bililive-go

多平台、多直播间录制工具，目前支持 哔哩哔哩，熊猫TV，战旗TV，斗鱼TV，火猫TV，龙珠直播，一直播，twitch   

![image](https://github.com/hr3lxphr6j/bililive-go/raw/master/screenshot.png)
## 依赖
* ffmpeg

## 下载
[releases](https://github.com/hr3lxphr6j/bililive-go/releases)

## 例子
```
./bililive-go -i "https://www.douyu.com/6655|https://www.panda.tv/10300"
```

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