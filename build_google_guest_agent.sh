#!/usr/bin/env bash

# Please note that the code below is modified by YANDEX LLC



GOOS="linux"
GOARCH="amd64"
if [[ $# -ne 1 ]]; then
  VERSION="$(git rev-parse --short HEAD)-yandex-patch-$(date '+%Y-%m-%d')"
else
  VERSION=${1}
fi

echo "VERSION: ${VERSION}"


CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o ./out/google_guest_agent ./google_guest_agent/
tar -czf ./out/google_guest_agent-${VERSION}.${GOOS}-${GOARCH}.tar.gz -C ./out ./google_guest_agent
