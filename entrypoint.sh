#!/bin/bash

chown -R ${PUID}:${PGID} /opt/bililive/
chown -R ${PUID}:${PGID} /srv/bililive/

umask ${UMASK}

exec su-exec ${PUID}:${PGID} /usr/bin/bililive-go -c /etc/bililive-go/config.yml
