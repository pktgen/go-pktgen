// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2019-2023 Intel Corporation

package l2p

import (
	"fmt"
)

type L2p struct {
}

func New() *L2p {
	fmt.Printf("Initializing L2P mode\n")

	return &L2p{}
}
