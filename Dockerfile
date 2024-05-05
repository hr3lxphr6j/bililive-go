FROM alpine

ARG tag

ENV WORKDIR="/srv/bililive" \
    OUTPUT_DIR="/srv/bililive" \
    CONF_DIR="/etc/bililive-go" \
    PORT=8080 \
    PUID=0 \
    PGID=0 \
    UMASK=022

RUN mkdir -p $OUTPUT_DIR $CONF_DIR && \
    apk add --no-cache ffmpeg libc6-compat curl su-exec tzdata && \
    cp -r -f /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

RUN set -ex; \
    case $(arch) in \
        aarch64) go_arch="arm64" ;; \
        arm*) go_arch="arm" ;; \
        i386|i686) go_arch="386" ;; \
        x86_64) go_arch="amd64" ;; \
    esac; \
    curl -fsSL -o bililive-linux-${go_arch}.tar.gz https://github.com/hr3lxphr6j/bililive-go/releases/download/${tag}/bililive-linux-${go_arch}.tar.gz; \
    if [ ! -f bililive-linux-${go_arch}.tar.gz ]; then \
        echo "Failed to download bililive-linux-${go_arch}.tar.gz"; \
        exit 1; \
    fi; \
    tar zxvf bililive-linux-${go_arch}.tar.gz; \
    if [ ! -f bililive-linux-${go_arch} ]; then \
        echo "The tar.gz did not contain bililive-linux-${go_arch}"; \
        exit 1; \
    fi; \
    chmod +x bililive-linux-${go_arch}; \
    mv bililive-linux-${go_arch} /usr/bin/bililive-go; \
    rm bililive-linux-${go_arch}.tar.gz

COPY config.yml $CONF_DIR/config.yml
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

VOLUME $OUTPUT_DIR

EXPOSE $PORT

WORKDIR ${WORKDIR}

ENTRYPOINT [ "sh" ]
CMD [ "/entrypoint.sh" ]