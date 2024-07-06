#!/bin/bash

export GO111MODULE=on
export GOPROXY=https://goproxy.io,direct
export CGO_ENABLED=1
export CGO_LDFLAGS="-lpcsclite -lnfc -lusb -lcurses"

go build \
    --ldflags "-linkmode external -extldflags -static -s -w" \
    -o _build/mister_arm/tapto.sh ./cmd/mister
