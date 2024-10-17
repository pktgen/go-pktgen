# SPDX-License-Identifier: BSD-3-Clause
# Copyright(c) 2023-2024 Intel Corporation

build:FORCE
	@go build -o ../../bin/iobind -gcflags '-N -l' || exit 1

rebuild:FORCE
	@rm -f ../../bin/iobind
	@go build -o ../../bin/iobind -gcflags '-N -l' || exit 1

FORCE:
