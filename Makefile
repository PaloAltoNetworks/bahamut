MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash -o pipefail

PROJECT_SHA ?= $(shell git rev-parse HEAD)
PROJECT_VERSION ?= $(lastword $(shell git tag --sort version:refname --merged $(shell git rev-parse --abbrev-ref HEAD)))
PROJECT_RELEASE ?= dev

ci: init lint test codecov

init:
	go get -u github.com/aporeto-inc/go-bindata/...
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	dep ensure
	dep status

lint:
	# --enable=unparam
	golangci-lint run \
		--disable-all \
		--exclude-use-default=false \
		--enable=errcheck \
		--enable=goimports \
		--enable=ineffassign \
		--enable=golint \
		--enable=unused \
		--enable=structcheck \
		--enable=staticcheck \
		--enable=varcheck \
		--enable=deadcode \
		--enable=unconvert \
		--enable=misspell \
		--enable=prealloc \
		--enable=nakedret \
		./...

test:
	@ go test ./... -race -cover -covermode=atomic -coverprofile=unit_coverage.cov

coverage_aggregate:
	@ mkdir -p artifacts
	@ for f in `find . -maxdepth 1 -name '*.cov' -type f`; do \
		filename="$${f##*/}" && \
		go tool cover -html=$$f -o artifacts/$${filename%.*}.html; \
	done;

codecov: coverage_aggregate
	bash <(curl -s https://codecov.io/bash)
