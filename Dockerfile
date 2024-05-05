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
    apk update && \
    apk add --no-cache ffmpeg libc6-compat curl su-exec tzdata && \
    cp -r -f /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    rm -rf /var/cache/apk/*

RUN set -ex; \
    case $(arch) in \
        aarch64) go_arch="arm64" ;; \
        arm*) go_arch="arm" ;; \
        i386|i686) go_arch="386" ;; \
        x86_64) go_arch="amd64" ;; \
    esac; \
    curl -fsSLo bililive-linux-${go_arch}.tar.gz "https://github.com/hr3lxphr6j/bililive-go/releases/download/${tag}/bililive-linux-${go_arch}.tar.gz" && \
    tar zxvf bililive-linux-${go_arch}.tar.gz && \
    chmod +x bililive-linux-${go_arch} && \
    mv bililive-linux-${go_arch} /usr/bin/bililive-go && \
    rm -rf bililive-linux-${go_arch}.tar.gz /tmp/*

COPY config.yml /defaults/config.yml
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

VOLUME $OUTPUT_DIR
EXPOSE $PORT

WORKDIR ${WORKDIR}
ENTRYPOINT [ "sh" ]
CMD [ "/entrypoint.sh" ]