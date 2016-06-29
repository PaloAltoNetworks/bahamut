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

GITHUB_TOKEN?=

## Update

domingoupdate:
	@echo "# Running domingoupdate in" $(PWD)
	@echo "REMINDER: you need to export GITHUB_TOKEN for this to work"
	@curl --fail -o domingo.mk -H "Cache-Control: no-cache" -H "Authorization: token $(GITHUB_TOKEN)" https://raw.githubusercontent.com/aporeto-inc/domingo/master/domingo.mk
	@echo "domingo.mk updated!"

## Dependencies

domingoinit:
	@$(foreach dir,$(DIRS_WITH_MAKEFILES),pushd $(dir) && make domingoinit && popd;)
	@echo "# Running domingoinit in" $(PWD)
	go get ./...
	@if [ -f glide.lock ]; then glide install && glide update; fi


## Testing

domingotest: domingolint domingomock
	@$(foreach dir,$(DIRS_WITH_MAKEFILES),pushd $(dir) && make domingotest && popd;)
	@echo "# Running test in" $(PWD)
	[ -z "${TEST_DIRS}" ] || go vet ${TEST_DIRS}
	[ -z "${TEST_DIRS}" ] || go test -v -race -cover ${TEST_DIRS}

domingolint:
	@$(foreach dir,$(DIRS_WITH_MAKEFILES),pushd $(dir) && make domingolint && popd;)
	@echo "# Running lint in" $(PWD)
	golint .

domingomock: domingocleanmock domingocleanvendor
	@$(foreach dir,$(DIRS_WITH_MAKEFILES),pushd $(dir) && make domingomock && popd;)
	@echo "# Running domingomock in" $(PWD)
	rm -rf ${APOMOCK_FOLDER}
	mkdir -p ${APOMOCK_FOLDER}
	touch ${APOMOCK_FOLDER}/domingomock.log
	kennebec --package="${APOMOCK_PACKAGES}" --output-dir=${APOMOCK_FOLDER} -v=4 -logtostderr=true>>${PWD}/${APOMOCK_FOLDER}/domingomock.log 2>&1
	@if [ ! -d vendor ]; then mkdir vendor; fi;
	@if [ -d ${APOMOCK_FOLDER} ]; then cp -r ${APOMOCK_FOLDER}/* vendor; fi;


## Cleaning

domingocleanvendor:
	@$(foreach dir,$(DIRS_WITH_MAKEFILES),pushd $(dir) && make domingocleanvendor && popd;)
	@echo "# Running domingocleanvendor in" $(PWD)
	rm -rf vendor

domingocleanmock:
	@$(foreach dir,$(DIRS_WITH_MAKEFILES),pushd $(dir) && make domingocleanmock && popd;)
	@echo "# Running domingocleanmock in" $(PWD)
	rm -rf ${APOMOCK_FOLDER}


## Docker Test Container
PROJECT_OWNER?=github.com/aporeto-inc
PROJECT_NAME?=my-super-project
BUILD_NUMBER?=latest

define DOCKER_FILE
FROM 926088932149.dkr.ecr.us-west-2.amazonaws.com/domingo
MAINTAINER Antoine Mercadal <antoine@aporeto.com>
ADD . /go/src/$(PROJECT_OWNER)/$(PROJECT_NAME)
WORKDIR /go/src/$(PROJECT_OWNER)/$(PROJECT_NAME)
endef
export DOCKER_FILE


create_build_container:
	echo "$$DOCKER_FILE" > .dockerfile-test
	docker build --file .dockerfile-test -t $(PROJECT_NAME)-build-image:$(BUILD_NUMBER) .

run_build_container:
	docker run --rm $(PROJECT_NAME)-build-image:$(BUILD_NUMBER)

clean_build_container:
	docker rmi $(PROJECT_NAME)-build-image:$(BUILD_NUMBER)
	rm -f .dockerfile-test
