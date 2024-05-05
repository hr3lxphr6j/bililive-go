#!/bin/sh

HOME=/srv/bililive
CONFIG_PATH=/etc/bililive-go/config.yml

chown -R ${PUID}:${PGID} ${HOME}
umask ${UMASK}

exec su-exec ${PUID}:${PGID} /usr/bin/bililive-go -c $CONFIG_PATH