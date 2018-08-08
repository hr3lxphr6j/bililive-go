# Bililive-go
[![Build Status](https://travis-ci.org/hr3lxphr6j/bililive-go.svg?branch=master)](https://travis-ci.org/hr3lxphr6j/bililive-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/hr3lxphr6j/bililive-go)](https://goreportcard.com/report/github.com/hr3lxphr6j/bililive-go)
[![Github release](https://img.shields.io/github/release/hr3lxphr6j/bililive-go.svg)](https://github.com/hr3lxphr6j/bililive-go/releases/latest)
[![Docker Pulls](https://img.shields.io/docker/pulls/chigusa/bililive-go.svg)](https://hub.docker.com/r/chigusa/bililive-go/)
[![Bilibili](https://img.shields.io/badge/%E5%93%94%E5%93%A9%E5%93%94%E5%93%A9-%E6%9C%AA%E6%9D%A5%E7%A7%91%E6%8A%80%E7%8E%8B%E8%80%81%E8%8F%8A%E5%BD%95%E6%92%AD%E7%BB%84-ebb8d0.svg)](https://space.bilibili.com/18578203/)

Bililive-go是一个跨平台、多直播间录制工具    

![image](https://github.com/hr3lxphr6j/bililive-go/raw/master/docs/screenshot.png)


## 例子
- 本地
    ```
    ./bililive-go -i https://www.panda.tv/10300 -i https://www.douyu.com/6655
    ```
- docker
    ```
    docker run -v ~/Movies:/srv/bililive --rm chigusa/bililive-go -o /srv/bililive -i https://www.panda.tv/10300
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
```
$ go get github.com/hr3lxphr6j/bililive-go
$ $GOPATH/src/github.com/hr3lxphr6j/bililive-go
$ make
```

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