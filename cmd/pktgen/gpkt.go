// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2023-2024 Intel Corporation

package main

import (
	"fmt"
	"os"

	"github.com/ebitengine/purego"
	"github.com/pktgen/go-pktgen/internal/cfg"
	"github.com/pktgen/go-pktgen/internal/tlog"
)

const (
	apiLibName  string = "./c-lib/usr/local/lib/x86_64-linux-gnu/libgpkt_api.so"
	hmapLibName  string = "./c-lib/usr/local/lib/x86_64-linux-gnu/libgpkt_hmap.so"
	modeLibName string = "./c-lib/usr/local/lib/x86_64-linux-gnu/libgpkt_modes.so"
	tlogLibName string = "./c-lib/usr/local/lib/x86_64-linux-gnu/libgpkt_tlog.so"
)

type gpktLib struct {
	Name   string  // Name of the library.
	Handle uintptr // library handle from dlopen().
}

type gpktApi struct {
	FuncName string      // Function Name of the library.
	FuncPtr  interface{} // Function pointer.
}

type gpktApis struct {
	tlogSetPath func(path string) int // Set tlog path variable.
	gpktAddArgv func(arg string) int  // Add a argv value.
	gpktStart   func() int            // Start function.
	gpktStop    func() int            // Stop function.
}

var (
	gLibMap  map[string]*gpktLib // Map of library names to their handles.
	gLibList []gpktLib           // List of library names and their handles.
	gApi     gpktApis            // Set of C API functions and their pointers.
)

func openLibrary(name string) (uintptr, error) {
	return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}

func gpktLoadLibraries() error {

	// Initialize the library map and list plus the C API functions.
	gLibMap = make(map[string]*gpktLib)
	gLibList = []gpktLib{}
	gApi = gpktApis{}

	libList := []string{ // Library must be loaded in this order and new libraries should be added here.
		tlogLibName,
		hmapLibName,
		modeLibName,
		apiLibName,
	}

	// Load the libraries in the order they are listed above.
	for _, libName := range libList {
		gLib := gpktLib{Name: libName}
		gLibList = append(gLibList, gLib)
		gLibMap[libName] = &gLib

		if handle, err := openLibrary(libName); err != nil {
			return fmt.Errorf("error loading %s: %v", libName, err)
		} else {
			tlog.DoPrintf("Library %s loaded successfully\n", libName)
			gLib.Handle = handle
		}
	}

	// Add new API functions to be registered here.
	gpktApis := []gpktApi{
		{FuncName: "tlog_set_path", FuncPtr: &gApi.tlogSetPath},
		{FuncName: "gpkt_add_argv", FuncPtr: &gApi.gpktAddArgv},
		{FuncName: "gpkt_start", FuncPtr: &gApi.gpktStart},
		{FuncName: "gpkt_stop", FuncPtr: &gApi.gpktStop},
	}
	tlog.DoPrintf("Registering API functions...\n")
	for _, v := range gpktApis {
		tlog.DoPrintf("   %s\n", v.FuncName)

		// Use uintptr(0) for library handler pointer to use running program symbols.
		purego.RegisterLibFunc(v.FuncPtr, uintptr(0), v.FuncName)
	}
	return nil
}

// gpktApiStart returning the basic information string
func gpktApiStart(cfg *cfg.System) error {

	// Create the DPDK arguments
	argv, err := cfg.GetArgsDPDK()
	if err != nil {
		tlog.DoPrintf("error cfg.Parse() failed: %v\n", err)
		return err
	}
	fmt.Printf("Starting Go-Pktgen with argc %d: %v\n", len(argv), argv)

	argc := len(argv)
	if argc == 0 {
		return fmt.Errorf("no configuration arguments found")
	}

	if err := gpktLoadLibraries(); err != nil {
		return fmt.Errorf("error loading Go-Pktgen APIs: %v", err)
	}

	// Set the Ptty if provided for logging
	if err := gApi.tlogSetPath(fmt.Sprintf("/dev/pts/%d", cfg.DebugTTY())); err < 0 {
		return fmt.Errorf("error setting Ptty: %s", err)
	}

	// Set the C-compatible array of strings
	for _, s := range argv {
		if ret := gApi.gpktAddArgv(s); ret < 0 {
			return fmt.Errorf("error setting C-compatible argument")
		}
	}

	// Clear the screen by scrolling the terminal to the top left corner
	for i := 0; i < 128; i++ {
		fmt.Fprintf(os.Stderr, "\n")
	}

	if cfg.DebugTTY() > 0 {
		tlog.DoPrintf("Starting Go-Pktgen with Ptty: /dev/pts/%d\n", cfg.DebugTTY())
	}

	// Initialize the DPDK system
	if ret := gApi.gpktStart(); ret < 0 {
		return fmt.Errorf("failed to initialize DPDK (%d)", ret)
	}

	return nil
}

// gpktApiStop is the function to stop DPDK
func gpktApiStop() {
	gApi.gpktStop()
}
