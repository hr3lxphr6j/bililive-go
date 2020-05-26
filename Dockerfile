FROM golang:1.14-alpine AS BUILD

ARG tag

RUN apk update && \
    apk add git make bash && \
    go get github.com/rakyll/statik && \
    go get github.com/golang/mock/mockgen && \
    mkdir -p /go/src/github.com/hr3lxphr6j/bililive-go && \
    git clone -b $tag --depth 1 https://github.com/hr3lxphr6j/bililive-go.git /go/src/github.com/hr3lxphr6j/bililive-go && \
    cd /go/src/github.com/hr3lxphr6j/bililive-go && \
    make bililive && \
    mv bin/bililive-linux-* bin/bililive-go

FROM alpine

ENV OUTPUT_DIR="/srv/bililive" \
    CONF_DIR="/etc/bililive-go" \
    PORT=8080

EXPOSE $PORT

RUN mkdir -p $OUTPUT_DIR && \
    mkdir -p $CONF_DIR && \
    apk update && \
    apk upgrade && \
    apk --no-cache add ffmpeg libc6-compat curl bash tree tzdata && \
    cp -r -f /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

VOLUME $OUTPUT_DIR

COPY --from=BUILD /go/src/github.com/hr3lxphr6j/bililive-go/bin/bililive-go /usr/bin/bililive-go
ADD config.docker.yml $CONF_DIR/config.yml

ENTRYPOINT ["/usr/bin/bililive-go"]
CMD ["-c", "/etc/bililive-go/config.yml"]
