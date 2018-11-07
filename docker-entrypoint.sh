#!/bin/bash
set -e
apt-get update
apt-get install -y dh-autoreconf libc6-dev-i386
git clone --branch jq-1.6 https://github.com/stedolan/jq /tmp/jq
pushd /tmp/jq
git submodule update --init
autoreconf -fi
./configure --with-oniguruma=builtin --disable-maintainer-mode
make LDFLAGS=-all-static
make install
ldconfig
wget https://github.com/goreleaser/goreleaser/releases/download/v0.93.0/goreleaser_amd64.deb
dpkg -i goreleaser_amd64.deb

popd

if [ $# -ge 1 ]; then
    if [ $1 == "release" ]; then
        goreleaser release --rm-dist
    else
        goreleaser release --skip-publish --skip-validate --rm-dist
    fi
fi