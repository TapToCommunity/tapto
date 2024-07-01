#!/bin/bash

GO111MODULE=on GOPROXY=https://goproxy.io,direct CGO_ENABLED=1 CGO_LDFLAGS="-lnfc -lusb -lcurses" go build \
    --ldflags "-linkmode external -extldflags -static -s -w" \
    -o _build/mister_arm/tapto.sh ./cmd/mister
