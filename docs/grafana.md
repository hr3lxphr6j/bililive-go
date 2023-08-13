# 三合一简化安装，原本的文档太晦涩了。

1. clone repo
>$ git clone https://github.com/hr3lxphr6j/bililive-go.git

2. 编辑 .env
默认账户密码 admin admin
不改网页里也会提醒你去改。

3. 命令行输入
>$ docker compose up

假如docker 版本低于20，需要安装docker-compose。新的版本是直接内置的

4. 浏览器打开 http://localhost:3000 。

tips
- 使用默认端口和别的端口冲突时，修改相关ports:
- `./Videos:/srv/bililive` 默认保存路径需要自定义


bibliography
1. [Docker-Compose-Prometheus-and-Grafana](https://github.com/Einsteinish/Docker-Compose-Prometheus-and-Grafana) 
