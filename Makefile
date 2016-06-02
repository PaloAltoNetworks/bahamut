# ------------------------------------------------
# Copyright (C) 2016 Aporeto Inc.
#
# File  : Makefile
#
# Author: alex@aporeto.com
# Date  : 2016-03-7
#
# ------------------------------------------------

MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash

.DEFAULT_GOAL := default

APOMOCK := .apomock
NOVENDOR := $(shell glide novendor)

DIRS_WITH_MAKEFILES := $(sort $(dir $(wildcard */Makefile)))

# Remove directories which have Makefiles from being tested by top level
NOTEST_DIRS := $(DIRS_WITH_MAKEFILES)
NOTEST_DIRS := $(addsuffix ...,$(NOTEST_DIRS))
NOTEST_DIRS := $(addprefix ./,$(NOTEST_DIRS))
TEST_DIRS := $(filter-out $(NOTEST_DIRS),$(NOVENDOR))

# Remove directories which are mock directories from being tested by top level
MOCK_DIRS := $(sort $(dir $(wildcard ./mock*)))
MOCK_DIRS := $(addsuffix ...,$(MOCK_DIRS))
MOCK_DIRS := $(addprefix ./,$(MOCK_DIRS))
TEST_DIRS := $(filter-out $(MOCK_DIRS),$(TEST_DIRS))

# Go Source
GO_SRCS := $(wildcard *.go)

.PHONY:  *

default: test, install

all: test install

clean:
	rm -rf test_results/*
	rm -f coverage.out

clean_vendor:
	rm -rf vendor

clean_apomock:
	rm -rf ${APOMOCK}

lint:
	@$(foreach dir,$(TEST_DIRS),golint $(dir);)
	golint .

test_packages:
	@echo "Test: Process directories with Makefiles" $(DIRS_WITH_MAKEFILES)
	@$(foreach dir,$(DIRS_WITH_MAKEFILES),echo "Test" $(dir) && pushd $(dir) && make test && make clean_vendor && popd;)

test: clean clean_vendor lint test_packages install_mock
	@echo "Test: Process directories without Makefiles" $(TEST_DIRS)
	[ -z "${TEST_DIRS}" ] || go vet ${TEST_DIRS}
	[ -z "${TEST_DIRS}" ] || go test -v -race -cover ${TEST_DIRS}
	@echo "Test: Skip mock directories" $(MOCK_DIRS)

convey: install_mock clean
	goconvey .

get:
	go get ./...

build: clean clean_vendor install_glide
	[ -z "${GO_SRCS}" ] || go build

install: clean clean_vendor install_glide
	[ -z "${GO_SRCS}" ] || go install

install_glide:
	@if [ -f glide.lock ]; then glide install; fi

update_glide:
	@if [ -f glide.lock ]; then glide update; fi

install_dependencies: clean_vendor get install_glide
	@echo "Install: Process directories with Makefiles" $(DIRS_WITH_MAKEFILES)
	@$(foreach dir,$(DIRS_WITH_MAKEFILES),echo "Install Dependencies in" $(dir) && pushd $(dir) && make install_dependencies && popd;)
	go get -u github.com/smartystreets/goconvey/convey
	go get -u github.com/aporeto-inc/kennebec/apomock
	go get -u github.com/golang/mock/gomock

init_glide:
	glide init
	glide install

MOCK_PACKAGES =

install_mock:
	@if [ ! -d vendor ]; then mkdir vendor; fi
	@if [ -d ${APOMOCK} ]; then cp -r ${APOMOCK}/* vendor; fi

apomock: clean_apomock install
	mkdir -p ${APOMOCK}
	touch ${APOMOCK}/apomock.log
	@echo "Mockgen: "
	kennebec --package="${MOCK_PACKAGES}" --output-dir=${APOMOCK} -v=4 -logtostderr=true>>${PWD}/${APOMOCK}/apomock.log 2>&1
