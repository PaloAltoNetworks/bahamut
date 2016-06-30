# ------------------------------------------------
# Copyright (C) 2016 Aporeto Inc.
#
# File  : Makefile
#
# Author: alex@aporeto.com, antoine@aporeto.com
# Date  : 2016-03-8
#
# ------------------------------------------------

MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash

APOMOCK_FOLDER := .apomock
APOMOCK_PACKAGES := $(shell if [ -f .apo.mock ]; then cat .apo.mock; fi)
NOVENDOR := $(shell glide novendor)

DIRS_WITH_MAKEFILES := $(sort $(dir $(wildcard */Makefile)))

# Remove directories which have Makefiles from being tested by top level
NOTEST_DIRS := $(DIRS_WITH_MAKEFILES)
NOTEST_DIRS := $(addsuffix ...,$(NOTEST_DIRS))
NOTEST_DIRS := $(addprefix ./,$(NOTEST_DIRS))

# Remove directories which are mock directories from being tested by top level
MOCK_DIRS := $(sort $(dir $(wildcard ./mock*)))
MOCK_DIRS := $(addsuffix ...,$(MOCK_DIRS))
MOCK_DIRS := $(addprefix ./,$(MOCK_DIRS))

TEST_DIRS := $(filter-out $(NOTEST_DIRS),$(NOVENDOR))
TEST_DIRS := $(filter-out $(MOCK_DIRS),$(TEST_DIRS))

PROJECT_OWNER?=github.com/aporeto-inc
PROJECT_NAME?=my-super-project
BUILD_NUMBER?=latest
GITHUB_TOKEN?=
DOCKER_REGISTRY?=926088932149.dkr.ecr.us-west-2.amazonaws.com
DOCKER_IMAGE_NAME?=$(PROJECT_NAME)
DOCKER_IMAGE_TAG?=$(BUILD_NUMBER)


## Update

domingo_update:
	@echo "# Updating Domingo..."
	@echo "REMINDER: you need to export GITHUB_TOKEN for this to work"
	@curl --fail -o domingo.mk -H "Cache-Control: no-cache" -H "Authorization: token $(GITHUB_TOKEN)" https://raw.githubusercontent.com/aporeto-inc/domingo/master/domingo.mk
	@echo "domingo.mk updated!"


## initialization

domingo_init:
	@$(foreach dir,$(DIRS_WITH_MAKEFILES),pushd $(dir) && make domingo_init && popd;)
	@echo "# Running domingo_init in" $(PWD)
	@if [ -f glide.lock ]; then glide install; else go get ./...; fi


## Testing

domingo_lint:
	@$(foreach dir,$(DIRS_WITH_MAKEFILES),pushd $(dir) && make domingo_lint && popd;)
	@echo "# Running lint in" $(PWD)
	golint .

domingo_test: domingo_lint
	@$(foreach dir,$(DIRS_WITH_MAKEFILES),pushd $(dir) && make domingo_test && popd;)
	@echo "# Running Domingo Tests in" $(PWD)
	@make domingo_save_vendor
	mkdir -p ${APOMOCK_FOLDER} vendor
	kennebec --package="${APOMOCK_PACKAGES}" --output-dir=${APOMOCK_FOLDER} -v=4 -logtostderr=true>>${PWD}/${APOMOCK_FOLDER}/apomock.log 2>&1
	@if [ -d ${APOMOCK_FOLDER} ]; then cp -r ${APOMOCK_FOLDER}/* vendor; fi;
	@echo "# Running test in" $(PWD)
	[ -z "${TEST_DIRS}" ] || go vet ${TEST_DIRS}
	[ -z "${TEST_DIRS}" ] || go test -v -race -cover ${TEST_DIRS}
	@make domingo_restore_vendor
	rm -rf ${APOMOCK_FOLDER}

domingo_save_vendor:
	@echo "# Saving vendor directory in" $(PWD)
	@if [ -d vendor ]; then cp -a vendor vendor.lock; fi

domingo_restore_vendor:
	@echo "# Restoring vendor directory in" $(PWD)
	@if [ -d vendor.lock ]; then rm -rf vendor && mv vendor.lock vendor; else rm -rf vendor; fi


## Docker Test Container

define DOCKER_FILE
FROM 926088932149.dkr.ecr.us-west-2.amazonaws.com/domingo
MAINTAINER Antoine Mercadal <antoine@aporeto.com>
ADD . /go/src/$(PROJECT_OWNER)/$(PROJECT_NAME)
WORKDIR /go/src/$(PROJECT_OWNER)/$(PROJECT_NAME)
endef
export DOCKER_FILE

domingo_contained_build:
	@echo "# Running domingo_build"
	echo "$$DOCKER_FILE" > .dockerfile-test
	mkdir -p /tmp/$(PROJECT_NAME)/$(BUILD_NUMBER)
	docker build --file .dockerfile-test -t $(PROJECT_NAME)-build-image:$(BUILD_NUMBER) .
	rm -f .dockerfile-test
	docker run --rm \
		-v /tmp/$(PROJECT_NAME)/$(BUILD_NUMBER):/export \
		-v /var/run/docker.sock:/var/run/docker.sock \
		$(PROJECT_NAME)-build-image:$(BUILD_NUMBER)
	docker rmi $(PROJECT_NAME)-build-image:$(BUILD_NUMBER)

domingo_docker_build:
	@echo "# Running domingo_docker_build"
	@if [ ! -f ./docker/Dockerfile ]; then echo "Error: No docker/Dockerfile!" && exit 1; fi;
	docker -H unix:///var/run/docker.sock build -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) docker

domingo_docker_push:
	@echo "# Running domingo_docker_push"
	docker -H unix:///var/run/docker.sock push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
	@echo "push!"
