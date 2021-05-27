MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash -o pipefail

export GO111MODULE = on

default: lint test

lint:
	@revive -config .revive.toml .

test:
	go test ./... -race -cover -covermode=atomic -coverprofile=unit_coverage.cov

sec:
	gosec -quiet ./...
