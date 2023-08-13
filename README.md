# Bililive-go
[![CI](https://github.com/hr3lxphr6j/bililive-go/actions/workflows/tests.yaml/badge.svg?branch=master)](https://github.com/hr3lxphr6j/bililive-go/actions/workflows/tests.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hr3lxphr6j/bililive-go)](https://goreportcard.com/report/github.com/hr3lxphr6j/bililive-go)
[![Github release](https://img.shields.io/github/release/hr3lxphr6j/bililive-go.svg)](https://github.com/hr3lxphr6j/bililive-go/releases/latest)
[![Docker Pulls](https://img.shields.io/docker/pulls/chigusa/bililive-go.svg)](https://hub.docker.com/r/chigusa/bililive-go/)
[![Bilibili](https://img.shields.io/badge/%E5%93%94%E5%93%A9%E5%93%94%E5%93%A9-%E6%9C%AA%E6%9D%A5%E7%A7%91%E6%8A%80%E7%8E%8B%E8%80%81%E8%8F%8A%E5%BD%95%E6%92%AD%E7%BB%84-ebb8d0.svg)](https://space.bilibili.com/18578203/)

Bililive-go是一个支持多种直播平台的直播录制工具   

![image](https://github.com/hr3lxphr6j/bililive-go/raw/master/docs/screenshot.webp)

## 支持网站

<table>
    <tr align="center">
        <th>站点</th>
        <th>url</th>
        <th>支持情况</th>
        <th>cookie</th>
    </tr>
    <tr align="center">
        <td>Acfun直播</td>
        <td>live.acfun.cn</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>哔哩哔哩直播</td>
        <td>live.bilibili.com</td>
        <td>滋瓷</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>战旗直播</td>
        <td>www.zhanqi.tv</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>斗鱼直播</td>
        <td>www.douyu.com</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>火猫直播</td>
        <td>www.huomao.com</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>龙珠直播</td>
        <td>longzhu.com</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>虎牙直播</td>
        <td>www.huya.com</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>CC直播</td>
        <td>cc.163.com</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>一直播</td>
        <td>www.yizhibo.com</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>twitch</td>
        <td>www.twitch.tv</td>
        <td>TODO</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>OPENREC</td>
        <td>www.openrec.tv</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>企鹅电竞</td>
        <td>egame.qq.com</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>浪live</td>
        <td>play.lang.live & www.lang.live</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>花椒</td>
        <td>www.huajiao.com</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>抖音直播</td>
        <td>live.douyin.com</td>
        <td>滋瓷</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>猫耳</td>
        <td>fm.missevan.com</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>克拉克拉</td>
        <td>www.hongdoufm.com</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>快手</td>
        <td>live.kuaishou.com</td>
        <td>滋瓷</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>YY直播</td>
        <td>www.yy.com</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
    <tr align="center">
        <td>微博直播</td>
        <td>weibo.com</td>
        <td>滋瓷</td>
        <td></td>
    </tr>
</table>

### cookie 在 config.yml 中的设置方法

cookie的设置以域名为单位。比如想在录制抖音直播时使用 cookie，那么 config.yml 中可以像下面这样写：
```
cookies:
  live.douyin.com: __ac_nonce=123456789012345678903;name=value
```

## Grafana 面板

> 请自行部署 prometheus 和 grafana

![image](https://github.com/hr3lxphr6j/bililive-go/raw/master/docs/dashboard.webp)

增加说明
[grafana](docs/grafana.md)

## 依赖
* [ffmpeg](https://ffmpeg.org/)

## 使用例子
- 本地
    ```
    ./bililive-go -i https://live.bilibili.com/1030 -i https://www.douyu.com/6655
    ```
- docker
    ```
    docker run --restart=always -v ~/Videos:/srv/bililive -p 8080:8080 -d chigusa/bililive-go
    ```

## 开发环境搭建（linux系统）
```
一、环境准备
  1. 前端环境
    1）前往https://nodejs.org/zh-cn/下载当前版本node（18.12.1）
    2）命令行运行 node -v 若控制台输出版本号则前端环境搭建成功
  2.后端环境
    1)下载golang安装 版本号1.19
      国际: https://golang.org/dl/
      国内: https://golang.google.cn/dl/
    2)命令行运行 go 若控制台输出各类提示命令 则安装成功 输入 go version 确认版本
  3.安装 ffmpeg (以centos7为例)
    1) yum install -y epel-release rpm
    2) rpm --import /etc/pki/rpm-gpg/RPM-GPG-KEY-EPEL-7
    3) yum repolist
    4) rpm --import http://li.nux.ro/download/nux/RPM-GPG-KEY-nux.ro
    5) rpm -Uvh http://li.nux.ro/download/nux/dextop/el7/x86_64/nux-dextop-release-0-1.el7.nux.noarch.rpm
    6) yum repolist
    7) yum install -y ffmpeg
二、克隆代码并编译(linux环境)    
   1. git clone https://github.com/hr3lxphr6j/bililive-go.git
   2. cd bililive-go
   3. make build-web
   4. make 
三、linux编译其他环境(以windows 为例)
   1. GOOS=windows GOARCH=amd64 CGO_ENABLED=0 UPX_ENABLE=0 TAGS=dev ./src/hack/build.sh bililive
   2.如果不需要调试，可以改成
      GOOS=windows GOARCH=amd64 CGO_ENABLED=0 UPX_ENABLE=0 TAGS=release ./src/hack/build.sh bililive
```

## Wiki
[Wiki](https://github.com/hr3lxphr6j/bililive-go/wiki)

## API
[API doc](https://github.com/hr3lxphr6j/bililive-go/blob/master/docs/API.md)

## 参考
- [you-get](https://github.com/soimort/you-get)
- [ykdl](https://github.com/zhangn1985/ykdl)
- [youtube-dl](https://github.com/ytdl-org/youtube-dl)
