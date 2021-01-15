# Build Frontend Start

# NOTE: Yarn has problems executing on ARM, so build on x86.
FROM --platform=linux/amd64 node:15.5.1-alpine as NODE_BUILD

ARG BUILDPLATFORM
ARG TARGETPLATFORM

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories && \
    apk update && \
    apk add git yarn make && \
    git clone https://github.com/hr3lxphr6j/bililive-go.git /bililive-go && \
    cd /bililive-go && \
    make build-web

# Build Frontend End

# Build Backend Start

FROM golang:1.15.6-alpine AS GO_BUILD

COPY --from=NODE_BUILD /bililive-go/ /go/src/github.com/hr3lxphr6j/bililive-go/

ENV GO111MODULE = ON \
    GOPROXY=https://goproxy.cn,direct

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories && \
    apk update && \
    apk add git make && \
    go get github.com/rakyll/statik && \
    go get github.com/golang/mock/mockgen && \
    cd /go/src/github.com/hr3lxphr6j/bililive-go && \
    make generate bililive && \
    mv bin/bililive-linux-* bin/bililive-go

# Build Backend End

# Build Runtime Image Start

FROM alpine

ENV OUTPUT_DIR="/srv/bililive" \
    CONF_DIR="/etc/bililive-go" \
    PORT=8080

EXPOSE $PORT

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories && \
    mkdir -p $OUTPUT_DIR && \
    mkdir -p $CONF_DIR && \
    apk update && \
    apk --no-cache add ffmpeg libc6-compat curl tzdata && \
    cp -r -f /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

VOLUME $OUTPUT_DIR

COPY --from=GO_BUILD /go/src/github.com/hr3lxphr6j/bililive-go/bin/bililive-go /usr/bin/bililive-go
ADD config.docker.yml $CONF_DIR/config.yml

ENTRYPOINT ["/usr/bin/bililive-go"]
CMD ["-c", "/etc/bililive-go/config.yml"]

# Build Runtime Image End
