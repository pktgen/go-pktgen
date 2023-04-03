# SPDX-License-Identifier: BSD-3-Clause
# Copyright(c) 2022-2024 Intel Corporation

build:FORCE
	@go build -o ../bin/pktgen -gcflags '-N -l' || exit 1

rebuild:FORCE
	@rm -f ../bin/serde
	@go build -o ../bin/pktgen -gcflags '-N -l' || exit 1

FORCE:
