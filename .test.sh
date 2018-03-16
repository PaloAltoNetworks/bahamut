#!/bin/bash
set -e

gometalinter \
    --exclude bindata.go \
    --exclude vendor \
    --vendor \
    --disable-all \
    --enable vet \
    --enable vetshadow \
    --enable golint \
    --enable ineffassign \
    --enable goconst \
    --enable errcheck \
    --enable varcheck \
    --enable structcheck \
    --enable gosimple \
    --enable misspell \
    --enable deadcode \
    --enable staticcheck \
    --deadline 5m \
    --tests ./...

for d in $(go list ./... | grep -v vendor); do
    go test -race -coverprofile=profile.out -covermode=atomic "$d"
    if [ -f profile.out ]; then cat profile.out >> coverage.txt; rm -f profile.out; fi
done
