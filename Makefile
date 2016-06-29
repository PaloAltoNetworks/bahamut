include domingo.mk

PROJECT_NAME := bahamut

ci: domingo_contained_build
init: domingo_init
test: domingo_test
release:
