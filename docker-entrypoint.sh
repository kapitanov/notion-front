#!/usr/bin/env sh

cd /opt/notion-front
/opt/notion-front/notion-front -i "$SOURCE_DIR" -c "$CACHE_DIR" -p "$LISTEN_ADDR"
