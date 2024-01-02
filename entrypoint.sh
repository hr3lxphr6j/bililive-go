#!/bin/sh

HOME=/srv/bililive

chown -R ${PUID}:${PGID} ${HOME}

umask ${UMASK}

exec su-exec ${PUID}:${PGID} /usr/bin/bililive-go -c /etc/bililive-go/config.yml
