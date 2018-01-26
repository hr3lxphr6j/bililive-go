FROM alpine

MAINTAINER chigusa

ENV OUTPUT_DIR /srv/bililive
ENV PORT 8080

RUN mkdir -p $OUTPUT_DIR && \
    apk update && \
    apk upgrade && \
    apk --no-cache add ffmpeg

EXPOSE $PORT
VOLUME $OUTPUT_DIR

ADD bililive /usr/bin

CMD bililive --enable-rpc --port $PORT -o $OUTPUT_DIR