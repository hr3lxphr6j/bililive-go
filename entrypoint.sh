#!/bin/sh

HOME=/srv/bililive

if [ ! -f $CONF_DIR/config.yml ]; then
    cp /defaults/config.yml $CONF_DIR/config.yml
fi

chown -R ${PUID}:${PGID} ${HOME}

umask ${UMASK}

exec su-exec ${PUID}:${PGID} /usr/bin/bililive-go -c /etc/bililive-go/config.yml