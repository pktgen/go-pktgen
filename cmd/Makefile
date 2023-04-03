# SPDX-License-Identifier: BSD-3-Clause
# Copyright (c) 2022-2024 Intel Corporation

build: FORCE
	make -C serde build
	make -C pktgen build

go-mod: FORCE
	(cd pktgen; go mod tidy)
	(cd serde; go mod tidy)

FORCE:
