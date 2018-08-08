FROM alpine

MAINTAINER chigusa


ENV OUTPUT_DIR="/srv/bililive" \
    CONF_DIR="/etc/bililive-go" \
    PORT=8080

EXPOSE $PORT

RUN mkdir -p $OUTPUT_DIR && \
    mkdir -p $CONF_DIR && \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk update && \
    apk upgrade && \
    apk --no-cache add ffmpeg libc6-compat curl bash tree tzdata && \
    cp -r -f /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

VOLUME $OUTPUT_DIR

ADD ./bin/bililive-go-linux-amd64 /usr/bin/bililive-go
ADD ./config.docker.yml $CONF_DIR/config.yml

ENTRYPOINT ["/usr/bin/bililive-go"]
CMD ["-c", "/etc/bililive-go/config.yml"]