# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2022-2024 Intel Corporation

#
# Head Makefile for compiling go-pktgen, but just a wrapper around
# meson and ninja using the scripts/gpkt-build.sh script.
#
# Use 'make' or 'make build' to build go-pktgen. If the build directory does
# not exist it will be created with these two build types.
#

mkfile_path=$(abspath $(lastword $(MAKEFILE_LIST)))
source_dir=$(shell dirname "$(mkfile_path)")
Build="${source_dir}/scripts/gpkt-build.sh"
Builder?=docker #OCI Image Builder
CE?=docker #Container Engine
OCI-Builder=$(shell echo $(Builder) | tr A-Z a-z)
ContainerEngine=$(shell echo $(CE) | tr A-Z a-z)

# Use V=1 on the make line to enable verbose output
ifeq ($V,1)
	verbose=-v
else
	verbose=
endif

all: rebuild-install

help: FORCE
	${Build} help
	@echo ""
	@echo "Makefile options:"
	@echo " Adding 'static_build=1' to the make line enables building static files"
	@echo "    eg: 'make static_build=1 rebuild install' for static executables"

rebuild: FORCE
	@echo "@@@@@ rebuild"
ifeq (${static_build},1)
	${Build} ${verbose} clean static build
else
	${Build} ${verbose} clean build
endif

build: FORCE
	@echo "@@@@@ build"
ifeq (${static_build},1)
	${Build} ${verbose} static build
else
	${Build} ${verbose} build
endif

rebuild-install: FORCE
	@echo "@@@@@ rebuild-install"
ifeq (${static_build},1)
	${Build} ${verbose} clean static build install
else
	${Build} ${verbose} clean build install
endif

coverity: FORCE
	@echo "@@@@@ coverity"
	${Build} ${verbose} clean coverity

debug: FORCE
	@echo "@@@@@ debug"
ifeq (${static_build},1)
	${Build} ${verbose} static debug
else
	${Build} ${verbose} debug
endif

debugopt: FORCE
	@echo "@@@@@ debugopt"
ifeq (${static_build},1)
	${Build} ${verbose} static debugopt
else
	${Build} ${verbose} debugopt
endif

clean: FORCE
	@echo "@@@@@ clean"
	${Build} ${verbose} clean

install: FORCE
	@echo "@@@@@ install"
	${Build} ${verbose} install

uninstall: FORCE
	@echo "@@@@@ uninstall"
	${Build} ${verbose} uninstall

docs: FORCE
	@echo "@@@@@ docs"
	${Build} ${verbose} docs

FORCE:
