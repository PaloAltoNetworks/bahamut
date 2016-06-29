include apobuild.mk

PROJECT_NAME := bahamut

clean: apoclean_vendor apoclean_apomock
init: apoinit
test: apotest
release:

ci: create_build_container run_build_container clean_build_container
