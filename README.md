# Bililive-go
[![Build Status](https://travis-ci.org/hr3lxphr6j/bililive-go.svg?branch=master)](https://travis-ci.org/hr3lxphr6j/bililive-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/hr3lxphr6j/bililive-go)](https://goreportcard.com/report/github.com/hr3lxphr6j/bililive-go)

Bililive-go是一个跨平台、多直播间录制工具   
By：[未来科技王老菊录播组](https://space.bilibili.com/18578203/)   

![image](https://github.com/hr3lxphr6j/bililive-go/raw/master/screenshot.webp)


## 例子
```
./bililive-go -i https://www.panda.tv/10300 -i https://www.douyu.com/6655
```

## 支持网站

<table>
    <tr align="center">
        <th>站点</th>
        <th>url</th>
        <th>支持情况</th>
    </tr>
    <tr align="center">
        <td>哔哩哔哩直播</td>
        <td>live.bilibili.com</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>熊猫直播</td>
        <td>www.panda.tv</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>战旗直播</td>
        <td>www.zhanqi.tv</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>斗鱼直播</td>
        <td>www.douyu.com</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>火猫直播</td>
        <td>www.huomao.com</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>龙珠直播</td>
        <td>longzhu.com</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>虎牙直播</td>
        <td>www.huya.com</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>全民直播</td>
        <td>www.quanmin.tv</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>CC直播</td>
        <td>cc.163.com</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>一直播</td>
        <td>www.yizhibo.com</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>twitch</td>
        <td>www.twitch.tv</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>OPENREC</td>
        <td>www.openrec.tv</td>
        <td>滋瓷</td>
    </tr>
</table>

## 依赖
* [ffmpeg](https://ffmpeg.org/)

## 下载&安装
* 安装ffmpeg
    * Windows：[FFmpeg Builds](https://ffmpeg.zeranoe.com/builds/)
    * macOS: [FFmpeg Builds](https://ffmpeg.zeranoe.com/builds/) or `brew install ffmpeg`
    * Linux: 从对应的包管理器上安装或从源码构建

* 下载Bililive-go 
    * [releases](https://github.com/hr3lxphr6j/bililive-go/releases)

## 获取&编译
`go get -v github.com/hr3lxphr6j/bililive-go`

## 使用
```
usage: BiliLive-go [<flags>]

A command-line live stream save tools.

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
      --version              Show application version.
      --debug                Enable debug mode.
  -t, --interval=20          Interval of query live status
  -o, --output="./"          Output file path.
  -i, --input=INPUT ...      Live room urls
  -c, --config=CONFIG        Config file.
      --enable-rpc           Enable RPC server.
      --rpc-addr=":8080"     RPC server listen port
      --rpc-token=RPC-TOKEN  RPC server token.
      --enable-rpc-tls       Enable TLS for RPC server
      --rpc-tls-cert-file=RPC-TLS-CERT-FILE  
                             Cert file for TLS on RPC
      --rpc-tls-key-file=RPC-TLS-KEY-FILE  
                             Key file for TLS on RPC

```

## 配置文件
```yaml
rpc: 
  enable: true            # 是否开启API
  port: 127.0.0.1:8080    # 监听地址
  token: ""               # token
  tls:                    # tls配置
    enable: false
    cert_file: ""
    key_file: ""
debug: false              # debug模式
interval: 15              # 直播间状态查询间隔时间（秒）
out_put_path: ./          # 输出文件路径
live_rooms:               # 直播间url
- https://www.panda.tv/10300
```

## API
[API doc](https://github.com/hr3lxphr6j/bililive-go/blob/master/docs/API.md)