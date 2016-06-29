include domingo.mk

PROJECT_NAME := bahamut

clean: domingocleanvendor domingocleanmock
init: domingoinit
test: domingotest
release:

ci: create_build_container run_build_container clean_build_container
