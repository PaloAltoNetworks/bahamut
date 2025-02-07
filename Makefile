MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash -o pipefail

export GO111MODULE = on

default: lint test

lint:
	# --enable=unparam
	golangci-lint run \
		--disable-all \
		--exclude-use-default=false \
		--exclude=package-comments \
		--exclude=unused-parameter \
		--enable=errcheck \
		--enable=goimports \
		--enable=ineffassign \
		--enable=revive \
		--enable=unused \
		--enable=staticcheck \
		--enable=unconvert \
		--enable=misspell \
		--enable=prealloc \
		--enable=nakedret \
		--enable=typecheck \
		./...
test:
	go test ./... -race -cover -covermode=atomic -coverprofile=unit_coverage.out

sec:
	gosec -quiet ./...

upgrade-deps:
	go get -u go.aporeto.io/elemental@master
	go get -u go.aporeto.io/tg@master
	go get -u go.aporeto.io/wsc@master

	go get -u github.com/nats-io/nats-server/v2@latest
	go get -u github.com/smartystreets/goconvey@latest
	go get -u go.uber.org/zap@latest

	go mod tidy
