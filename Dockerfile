FROM alpine

MAINTAINER chigusa


ENV OUTPUT_DIR /srv/bililive
ENV CONF_DIR /etc/bililive-go
ENV PORT 8080

EXPOSE $PORT

RUN mkdir -p $OUTPUT_DIR && \
    mkdir -p $CONF_DIR && \
    apk update && \
    apk upgrade && \
    apk --no-cache add ffmpeg

VOLUME $OUTPUT_DIR

ADD ./bin/bililive-go-linux-amd64 /usr/bin/bililive-go
ADD ./config.docker.yml $CONF_DIR/config.yml

ENTRYPOINT /usr/bin/bililive-go -c $CONF_DIR/config.yml