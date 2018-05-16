FROM alpine

MAINTAINER chigusa


ENV OUTPUT_DIR /srv/bililive
ENV PORT 8080

EXPOSE $PORT

RUN mkdir -p $OUTPUT_DIR && \
    apk update && \
    apk upgrade && \
    apk --no-cache add ffmpeg

EXPOSE $PORT
VOLUME $OUTPUT_DIR

ADD bililive-go /usr/bin

ENTRYPOINT /usr/bin/bililive-go --enable-rpc --rpc-addr :$PORT -o $OUTPUT_DIR