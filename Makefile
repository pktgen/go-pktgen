# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2022-2024 Intel Corporation

all: build-c-lib build FORCE

build-c-lib: FORCE
	@make -C c-lib rebuild-install # rebuild the C code library

build: FORCE
	@make -C cmd build

go-mod: FORCE
	@go clean --cache
	@go clean --modcache
	@make -C internal go-mod
	@make -C cmd go-mod

FORCE:
