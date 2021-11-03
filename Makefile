MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash -o pipefail

export GO111MODULE = on

default: lint test

lint:
	# --enable=unparam
	# --enable=typecheck this one causes a false positive on go 1.16
	golangci-lint run \
		--disable-all \
		--exclude-use-default=false \
		--enable=errcheck \
		--enable=goimports \
		--enable=ineffassign \
		--enable=revive \
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
	go test ./... -race -cover -covermode=atomic -coverprofile=unit_coverage.cov

	@ echo "Converting the coverage file..."
	gocov convert ./unit_coverage.cov | gocov-xml > ./coverage.xml

sec:
	gosec -quiet ./...
