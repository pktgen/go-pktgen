// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package main

/*
#include <stdio.h>
#include <stdlib.h>

#include <gpkt_api.h>
*/
import "C"

import (
	"fmt"
	"os"

	"github.com/pktgen/go-pktgen/internal/cfg"
	"github.com/pktgen/go-pktgen/internal/tlog"
)

// gpktApiStart returning the basic information string
func gpktApiStart(cs *cfg.System) error {

	// Convert the configData to C-compatible types
	argv, err := cs.MakeArgs()
	if err != nil {
		return err
	}

	argc := len(argv)
	if argc == 0 {
		return fmt.Errorf("no configuration arguments found")
	}

	// Set the C-compatible array of strings
	for _, s := range argv {
		cStr := C.CString(s)
		C.gpktSetArgv(cStr)
	}

	// Set the Ptty if provided for logging
	cStr := C.CString("")
	if pktgen.cfg.DebugTTY() > 0 {
		cStr = C.CString(fmt.Sprintf("%v", pktgen.cfg.DebugTTY()))
	}

	// Clear the screen by scrolling the terminal to the top left corner
	for i := 0; i < 128; i++ {
        fmt.Fprintf(os.Stderr, "\n");
    }

	// Initialize the DPDK
	if ret := C.gpktStart(cStr); ret < 0 {
		return fmt.Errorf("failed to initialize DPDK")
	}

	return nil
}

// gpktApiStop is the function to stop DPDK
func gpktApiStop() {
	tlog.DoPrintf("Calling gpktStop()\n")
    C.gpktStop()
	tlog.DoPrintf("Done calling gpktStop()\n")
}
