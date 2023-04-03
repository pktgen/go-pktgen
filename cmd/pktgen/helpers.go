// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package main

import (
	"fmt"
	"os"

	cz "github.com/pktgen/go-pktgen/internal/colorize"
)

// PktgenInfo returning the basic information string
func PktgenInfo(color bool) string {
	if !color {
		return fmt.Sprintf("%s, Version: %s Pid: %d %s",
			"Go-Pktgen powered by DPDK", Version(), os.Getpid(),
			"Copyright © 2022-2024 Intel Corporation")
	}

	return fmt.Sprintf("%s, Version: %s Pid: %s %s",
		cz.Yellow("Go-Pktgen Traffic Generator"), cz.Green(Version()),
		cz.Red(os.Getpid()),
		cz.SkyBlue("Copyright © 2022-2024 Intel Corporation"))
}
