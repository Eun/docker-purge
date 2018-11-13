#!/bin/sh
set -e
apk update
apk add build-base autoconf git automake libtool
git clone --branch jq-1.6 https://github.com/stedolan/jq /tmp/jq
WD=$(pwd)
cd /tmp/jq
git submodule update --init
autoreconf -fi
./configure --with-oniguruma=builtin --disable-maintainer-mode
make LDFLAGS=-all-static
make install
ldconfig || true

cd $WD


go build -o dist/alpine/docker-purge