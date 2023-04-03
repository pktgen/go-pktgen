# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2023-2025 Intel Corporation

mainVersion := $(shell cat ../../VERSION)
buildDate := $(shell date +"%c")
ld_flags := -ldflags "-X 'main.version=$(mainVersion)' -X 'main.buildDate=$(buildDate)'"

build:FORCE
	@go build -o ../../bin/iobind -gcflags '-N -l' $(ld_flags) || exit 1

rebuild:FORCE
	@rm -f ../../bin/iobind
	@go build -o ../../bin/iobind -gcflags '-N -l' || exit 1

FORCE:
