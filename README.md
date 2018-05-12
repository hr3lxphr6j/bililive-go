# Bililive-go

Bililive-go是一个跨平台、多直播间录制工具   
By：[未来科技王老菊录播组](https://space.bilibili.com/18578203/)   

![image](https://github.com/hr3lxphr6j/bililive-go/raw/master/screenshot.png)


## 例子
```
./bililive-go -i "https://www.panda.tv/10300|https://www.douyu.com/6655"
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
        <td>一直播</td>
        <td>www.yizhibo.com</td>
        <td>滋瓷</td>
    </tr>
    <tr align="center">
        <td>twitch</td>
        <td>www.twitch.tv</td>
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
Usage: bililive-go [-hv] [-i urls] [-o path] [-t seconds] [-c filename]
Options:
  -h: 显示帮助并退出
  -v: 显示版本并退出
  -i: 直播间地址，若有多个直播间请使用 "|" 进行分割
  -o: 输出文件目录，默认为 ./ ，及为当前目录 
  -t: 直播间状态查询间隔时间，默认30(s)
  -c: 设置配置文件，命令行的参数将会覆盖配置文件中相同的设置，默认读取 ./config.yml
```

## 配置文件
```yaml
rpc: 
  enable: true            # 是否开启API
  port: 127.0.0.1:8080    # 监听地址
  token: ""               # token,在header中传递
  tls:                    # tls配置
    enable: false
    cert_file: ""
    key_file: ""
log_level: info           # log等级(info|debug...)
interval: 15              # 直播间状态查询间隔时间（秒）
out_put_path: ./          # 输出文件路径
live_rooms:               # 直播间url
- https://www.panda.tv/10300
```

## API
[API doc](https://github.com/hr3lxphr6j/bililive-go/raw/rpc-b/API.md)