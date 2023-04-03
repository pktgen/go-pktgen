# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2023-2025 Intel Corporation

all: build-c-lib build FORCE

build-c-lib: FORCE
	@rm -f bin/*
	@make -C c-lib rebuild-install # rebuild the C code library

build: FORCE
	@make -C cmd build

go-mod: FORCE
	@go clean --cache
	@go clean --modcache
	@make -C pkgs go-mod
	@make -C internal go-mod
	@make -C cmd go-mod

FORCE:
