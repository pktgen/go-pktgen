// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package main

/*
#include <stdio.h>
#include <stdlib.h>
*/
//import "C"

import (
	"fmt"
	"os"

	"github.com/ebitengine/purego"
	"github.com/pktgen/go-pktgen/internal/cfg"
	"github.com/pktgen/go-pktgen/internal/tlog"
)

const (
	gpktApiLibName    string = "libgpkt_api.so"
	gpktSingleLibName string = "libgpkt_single.so"
)

type gpktLib struct {
	libName string  // Name of the library.
	libPtr  uintptr // library handle from dlopen().
}

type gpktApiLib struct {
	gpktLib

	gpktStart   func() int // Start function.
	gpktStop    func() int                        // Stop function.
	gpktSetArgv func(arg string) int              // Set the argv values function.
	tlogOpen    func(pts int) int                 // Open function for tlog.
}

type gpktSingleLib struct {
	gpktLib
}

var (
	gApi    gpktApiLib
	gSingle gpktSingleLib
)

func openLibrary(name string) (uintptr, error) {
	return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}

func gpktLoadApis() error {

	if lib, err := openLibrary(gpktApiLibName); err != nil {
		return fmt.Errorf("error loading %s", gpktApiLibName)
	} else {
		g := gpktApiLib{}
		g.libName = gpktApiLibName
		g.libPtr = lib

		purego.RegisterLibFunc(&g.gpktStart, uintptr(lib), "gpktStart")
		purego.RegisterLibFunc(&g.gpktStop, uintptr(lib), "gpktStop")
		purego.RegisterLibFunc(&g.gpktSetArgv, uintptr(lib), "gpktSetArgv")
		purego.RegisterLibFunc(&g.tlogOpen, uintptr(lib), "tlog_open")

		gApi = g
	}

	if lib, err := openLibrary(gpktSingleLibName); err != nil {
		return fmt.Errorf("error loading %s", gpktSingleLibName)
	} else {
		g := gpktSingleLib{}
		g.libName = gpktApiLibName
		g.libPtr = lib

		gSingle = g

	}
	return nil
}

// gpktApiStart returning the basic information string
func gpktApiStart(c *cfg.System) error {

	// Convert the configData to C-compatible types
	argv, err := c.MakeArgs()
	if err != nil {
		tlog.DoPrintf("error MakeArgs() failed: %v\n", err)
		return err
	}

	argc := len(argv)
	if argc == 0 {
		return fmt.Errorf("no configuration arguments found")
	}

	tlog.DoPrintf("Starting Go-Pktgen with argc %d: %v\n", argc, argv)

	if err := gpktLoadApis(); err != nil {
		return fmt.Errorf("error loading Go-Pktgen APIs: %v", err)
	}

	// Set the Ptty if provided for logging
	if err := gApi.tlogOpen(c.DebugTTY()); err < 0 {
		return fmt.Errorf("error setting Ptty: %d", err)
	}

	// Set the C-compatible array of strings
	for _, s := range argv {
		if ret := gApi.gpktSetArgv(s); ret < 0 {
            return fmt.Errorf("error setting C-compatible argument")
        }
	}

	// Clear the screen by scrolling the terminal to the top left corner
	for i := 0; i < 128; i++ {
		fmt.Fprintf(os.Stderr, "\n")
	}

	tlog.DoPrintf("Starting Go-Pktgen with Ptty: %v\n", c.DebugTTY())

	// Initialize the DPDK
	if ret := gApi.gpktStart(); ret < 0 {
		return fmt.Errorf("failed to initialize DPDK")
	}

	return nil
}

// gpktApiStop is the function to stop DPDK
func gpktApiStop() {
	gApi.gpktStop()
}
