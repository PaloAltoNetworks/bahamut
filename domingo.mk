# ------------------------------------------------
# Copyright (C) 2016 Aporeto Inc.
#
# File  : Makefile
#
# Author: alex@aporeto.com, antoine@aporeto.com
# Date  : 2016-03-8
#
# ------------------------------------------------

## configure this throught environment variables
PROJECT_OWNER?=github.com/aporeto-inc
PROJECT_NAME?=my-super-project
DOMINGO_DOCKER_TAG?=latest
DOMINGO_DOCKER_REPO=gcr.io/aporetodev
GITHUB_TOKEN?=

######################################################################
######################################################################

export ROOT_DIR?=$(PWD)

MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash -o pipefail

NOVENDOR                := $(shell glide novendor)
NOTEST_DIRS             := $(addsuffix ...,$(NOVENDOR))
NOTEST_DIRS             := $(addprefix ./,$(NOVENDOR))
TEST_DIRS               := $(filter-out $(NOTEST_DIRS),$(NOVENDOR))
GO_SRCS                 := $(wildcard *.go)

## Update

domingo_update:
	@echo "# Updating Domingo..."
	@echo "REMINDER: you need to export GITHUB_TOKEN for this to work"
	curl --fail -o domingo.mk -H "Cache-Control: no-cache" -H "Authorization: token $(GITHUB_TOKEN)" https://raw.githubusercontent.com/aporeto-inc/domingo/master/domingo.mk
	@echo "domingo.mk updated!"


## initialization

domingo_init:
	@if [ -f glide.yaml ]; then glide install; else go get ./...; fi

## Testing

domingo_goconvey:
	goconvey -port 34562 .

domingo_test:
	@echo "Running linters army..."
	@gometalinter --vendor --disable-all \
		--enable=vet \
		--enable=vetshadow \
		--enable=golint \
		--enable=ineffassign \
		--enable=goconst \
		--enable=errcheck \
		--enable=varcheck \
		--enable=structcheck \
		--enable=gosimple \
		--enable=misspell \
		--deadline 5m \
		--tests $(TEST_DIRS)
	@echo "Running unit tests..."
	@go test -race -cover $(TEST_DIRS)
	@echo "Success!"

container:
	make build_linux
	cd docker && docker build -t $(DOMINGO_DOCKER_REPO)/$(PROJECT_NAME):$(DOMINGO_DOCKER_TAG) .
