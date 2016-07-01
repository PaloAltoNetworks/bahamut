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
DOMINGO_BASE_IMAGE?=$(DOCKER_REGISTRY)/domingo:1.0-2

######################################################################
######################################################################

export ROOT_DIR?=$(PWD)

MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash

APOMOCK_FILE 			:= .apomock
APOMOCK_PACKAGES 	:= $(shell if [ -f $(APOMOCK_FILE) ]; then cat $(APOMOCK_FILE); fi)
NOVENDOR 					:= $(shell glide novendor)
MANAGED_DIRS 			:= $(sort $(dir $(wildcard */Makefile)))
MOCK_DIRS 				:= $(sort $(dir $(wildcard */.apo.mock)))
NOTEST_DIRS 			:= $(MANAGED_DIRS)
NOTEST_DIRS 			:= $(addsuffix ...,$(NOTEST_DIRS))
NOTEST_DIRS 			:= $(addprefix ./,$(NOTEST_DIRS))
TEST_DIRS 				:= $(filter-out $(NOTEST_DIRS),$(NOVENDOR))

## Update

domingo_update:
	@echo "# Updating Domingo..."
	@echo "REMINDER: you need to export GITHUB_TOKEN for this to work"
	curl --fail -o domingo.mk -H "Cache-Control: no-cache" -H "Authorization: token $(GITHUB_TOKEN)" https://raw.githubusercontent.com/aporeto-inc/domingo/master/domingo.mk
	@echo "domingo.mk updated!"


## initialization

domingo_init:
	@echo "# Running domingo_init in" $(PWD)
	@if [ -f glide.lock ]; then glide install; else go get ./...; fi


## Testing

domingo_test:
	echo > $(ROOT_DIR)/testresults.xml
	make domingo_lint domingo_apomock
	sed -i.bak -E 's/(<testsuites>|<\/testsuites>)//g' $(ROOT_DIR)/testresults.xml
	rm -rf $(ROOT_DIR)/testresults.xml.bak
	echo "<?xml version=\"1.0\" encoding=\"utf-8\"?><testsuites>" | cat - $(ROOT_DIR)/testresults.xml > $(ROOT_DIR)/.testresults.xml
	echo '</testsuites>' >> $(ROOT_DIR)/.testresults.xml
	rm -f $(ROOT_DIR)/testresults.xml
	mv $(ROOT_DIR)/.testresults.xml $(ROOT_DIR)/testresults.xml
	if [ -d /outside_world ]; then cp $(ROOT_DIR)/testresults.xml /outside_world; fi

domingo_lint:
	@echo "# Running lint & vet"
	golint $(NOVENDOR)
	go vet $(NOVENDOR)

domingo_apomock:
	@$(foreach dir,$(MANAGED_DIRS),pushd $(dir) && make domingo_apomock && popd;)
	@echo "# Running ApoMock in" $(dir)
	if [ -f $(APOMOCK_FILE) ]; then make domingo_init_apomock; fi;
	go test -v -race -cover $(TEST_DIRS) | tee >(go2xunit -fail | tail -n +2 >> $(ROOT_DIR)/testresults.xml)
	if [ -f $(APOMOCK_FILE) ]; then make domingo_deinit_apomock; fi;

domingo_init_apomock:
	@make domingo_save_vendor
	kennebec --package="$(APOMOCK_PACKAGES)" --output-dir=vendor -v=4 -logtostderr=true >> /dev/null 2>&1

domingo_deinit_apomock:
	@make domingo_restore_vendor

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
		-t \
		--rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(ROOT_DIR):/outside_world \
		-e BUILD_NUMBER="$(BUILD_NUMBER)" \
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
