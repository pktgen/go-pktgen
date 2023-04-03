# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2023-2025 Intel Corporation

build: FORCE
	make -C iobind build
	make -C serde build
	make -C pktgen build

go-mod: FORCE
	(cd pktgen; go mod tidy)
	(cd serde; go mod tidy)
	(cd iobind; go mod tidy)

FORCE:
