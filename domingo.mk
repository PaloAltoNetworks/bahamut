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

DOMINGO_BASE_IMAGE?=926088932149.dkr.ecr.us-west-2.amazonaws.com/domingo:1.0-1
PROJECT_OWNER?=github.com/aporeto-inc
PROJECT_NAME?=my-super-project
BUILD_NUMBER?=latest
GITHUB_TOKEN?=
DOCKER_LOGIN_COMMAND?=
DOCKER_REGISTRY?=926088932149.dkr.ecr.us-west-2.amazonaws.com
DOCKER_IMAGE_NAME?=$(PROJECT_NAME)
DOCKER_IMAGE_TAG?=$(BUILD_NUMBER)
DOCKER_ENABLE_BUILD?=0
DOCKER_ENABLE_PUSH?=0
DOCKER_ENABLE_RETAG?=0

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
	[ -z "${TEST_DIRS}" ] || go test -v -race -cover ${TEST_DIRS} #| tee >(go2xunit -fail -output ./testresults.xml)
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
FROM $(DOMINGO_BASE_IMAGE)
MAINTAINER Antoine Mercadal <antoine@aporeto.com>
ADD . /go/src/$(PROJECT_OWNER)/$(PROJECT_NAME)
WORKDIR /go/src/$(PROJECT_OWNER)/$(PROJECT_NAME)
endef
export DOCKER_FILE

domingo_contained_build:
	@echo "# Running domingo_build"
	echo "$$DOCKER_FILE" > .dockerfile-domingo
	eval $(DOCKER_LOGIN_COMMAND)
	docker pull $(DOMINGO_BASE_IMAGE)
	docker build --file .dockerfile-domingo -t $(PROJECT_NAME)-build-image:$(BUILD_NUMBER) .
	rm -f .dockerfile-domingo
	docker run \
		--rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-e DOCKER_LOGIN_COMMAND="$(DOCKER_LOGIN_COMMAND)" \
		-e BUILD_NUMBER="$(BUILD_NUMBER)" \
		-e DOCKER_ENABLE_BUILD="$(DOCKER_ENABLE_BUILD)" \
		-e DOCKER_ENABLE_PUSH="$(DOCKER_ENABLE_PUSH)" \
		-e DOCKER_IMAGE_NAME="$(DOCKER_IMAGE_NAME)" \
		-e DOCKER_IMAGE_TAG="$(DOCKER_IMAGE_TAG)" \
		-e DOCKER_REGISTRY="$(DOCKER_REGISTRY)" \
		-e GITHUB_TOKEN="$(DOCKER_ENABLE_BUILD)" \
		-e PROJECT_NAME="$(PROJECT_NAME)" \
		-e PROJECT_OWNER="$(PROJECT_OWNER)" \
		$(PROJECT_NAME)-build-image:$(BUILD_NUMBER)
	docker rmi $(PROJECT_NAME)-build-image:$(BUILD_NUMBER)

ifeq ($(DOCKER_ENABLE_BUILD),1)
domingo_docker_build:
	@echo "# Running domingo_docker_build"
	docker -H unix:///var/run/docker.sock \
		build \
		-t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) \
		docker
else
domingo_docker_build:
	@echo " - docker build not explicitely enabled: run 'export DOCKER_ENABLE_BUILD=1'"
endif

ifeq ($(DOCKER_ENABLE_PUSH),1)
domingo_docker_push:
	@echo "# Running domingo_docker_push"
	eval $(DOCKER_LOGIN_COMMAND)
	docker -H unix:///var/run/docker.sock \
		push \
		$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
else
domingo_docker_push:
	@echo " - docker push not explicitely enabled: run 'export DOCKER_ENABLE_PUSH=1'"
endif

ifeq ($(DOCKER_ENABLE_RETAG),1)
domingo_docker_retag:
	@echo "# Running domingo_docker_tag_latest"
	eval $(DOCKER_LOGIN_COMMAND);
	docker -H unix:///var/run/docker.sock \
		tag \
		$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) \
		$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):master_latest
	docker -H unix:///var/run/docker.sock \
		push \
		$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):master_latest
else
domingo_docker_retag:
	@echo " - docker retag not explicitely enabled: run 'export DOCKER_ENABLE_RETAG=1'"
endif
