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
BUILD_NUMBER?=latest
GITHUB_TOKEN?=
DOCKER_LOGIN_COMMAND?=
DOCKER_REGISTRY?=926088932149.dkr.ecr.us-west-2.amazonaws.com
DOCKER_IMAGE_NAME?=$(PROJECT_NAME)
DOCKER_IMAGE_TAG?=$(BUILD_NUMBER)
DOCKER_ENABLE_BUILD?=0
DOCKER_ENABLE_PUSH?=0
DOCKER_ENABLE_RETAG?=0
DOCKER_LATEST_TAG?=latest
DOMINGO_EXPORT_FOLDER?=/outside_world
DOMINGO_BASE_IMAGE?=$(DOCKER_REGISTRY)/domingo

######################################################################
######################################################################

export ROOT_DIR?=$(PWD)

MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash -o pipefail

APOMOCK_FILE            := .apomock
APOMOCK_PACKAGES        := $(shell if [ -f $(APOMOCK_FILE) ]; then cat $(APOMOCK_FILE); fi)
NOVENDOR                := $(shell glide novendor)
MANAGED_DIRS            := $(sort $(dir $(wildcard */Makefile)))
MOCK_DIRS               := $(sort $(dir $(wildcard */.apomock)))
NOTEST_DIRS             := $(MANAGED_DIRS)
NOTEST_DIRS             := $(addsuffix ...,$(NOTEST_DIRS))
NOTEST_DIRS             := $(addprefix ./,$(NOTEST_DIRS))
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
	@echo "# Running domingo_init in" $(PWD)
	@if [ -f glide.yml ]; then glide up; else go get ./...; fi
	@go get -u github.com/aporeto-inc/kennebec

## Testing

domingo_goconvey:
	make domingo_lint domingo_init_apomock
	goconvey .
	make domingo_deinit_apomock

domingo_test:
	@$(foreach dir,$(MANAGED_DIRS),pushd ${dir} > /dev/null && make domingo_test && popd > /dev/null;)
	@if [ -f $(APOMOCK_FILE) ]; then make domingo_init_apomock; fi
	@if [ "$(GO_SRCS)" != "" ]; then go test -race -cover $(TEST_DIRS) || exit 1; else echo "# Skipped as no go sources found"; fi
	@if [ -f $(APOMOCK_FILE) ]; then make domingo_deinit_apomock; fi


domingo_init_apomock:
	@make domingo_save_vendor
	@kennebec --package="$(APOMOCK_PACKAGES)" --output-dir=vendor -v=4 -logtostderr=true >> /dev/null 2>&1

domingo_deinit_apomock:
	@make domingo_restore_vendor

domingo_save_vendor:
	@if [ -d vendor ]; then cp -a vendor vendor.lock; fi

domingo_restore_vendor:
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
		--name $(PROJECT_NAME)-build-container_$(BUILD_NUMBER) \
		--privileged \
		--net host \
		-t \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /dev/shm:/dev/shm \
		-v $(ROOT_DIR):$(DOMINGO_EXPORT_FOLDER) \
		-e BUILD_NUMBER="$(BUILD_NUMBER)" \
		-e DOMINGO_EXPORT_FOLDER="$(DOMINGO_EXPORT_FOLDER)" \
		-e DOCKER_HOST="unix:///var/run/docker.sock" \
		-e DOCKER_LOGIN_COMMAND="$(DOCKER_LOGIN_COMMAND)" \
		-e DOCKER_ENABLE_BUILD="$(DOCKER_ENABLE_BUILD)" \
		-e DOCKER_ENABLE_PUSH="$(DOCKER_ENABLE_PUSH)" \
		-e DOCKER_IMAGE_NAME="$(DOCKER_IMAGE_NAME)" \
		-e DOCKER_IMAGE_TAG="$(DOCKER_IMAGE_TAG)" \
		-e DOCKER_REGISTRY="$(DOCKER_REGISTRY)" \
		-e GITHUB_TOKEN="$(GITHUB_TOKEN)" \
		-e PROJECT_NAME="$(PROJECT_NAME)" \
		-e PROJECT_OWNER="$(PROJECT_OWNER)" \
		$(PROJECT_NAME)-build-image:$(BUILD_NUMBER)
	docker rm -f $(PROJECT_NAME)-build-container_$(BUILD_NUMBER)
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
		$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_LATEST_TAG)
	docker -H unix:///var/run/docker.sock \
		push \
		$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_LATEST_TAG)
else
domingo_docker_retag:
	@echo " - docker retag not explicitely enabled: run 'export DOCKER_ENABLE_RETAG=1'"
endif
