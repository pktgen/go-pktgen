// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package main

import (
	"fmt"
	"os"
	"syscall"

	hlp "github.com/pktgen/go-pktgen/internal/helpers"
	"github.com/pktgen/go-pktgen/internal/iobind"

	flags "github.com/jessevdk/go-flags"
)

type frame struct {
	data string
}

// IOBindTool to convert a text file to a packet file.
type IOBindTool struct {
	version   string // Version of tool
	buildDate string // Build date of tool
}

// Options command line options
type Options struct {
	BindDev     bool `short:"b" long:"bind" description:"Bind device to a driver"`
	UnbindDev   bool `short:"u" long:"unbind" description:"Unbind device from a driver"`
	ShowVersion bool `short:"v" long:"version" description:"Print out version and exit"`
	Verbose     bool `short:"V" long:"verbose" description:"Enable verbose output"`
}

// Global to the main package for the tool
var (
	version    string      // Version of tool
	buildDate  string      // Build date of tool
	iobindTool *IOBindTool // Global IOBindTool instance
	options    Options     // Command line options
	parser     = flags.NewParser(&options, flags.Default)
)

// Setup the tool's global information and startup the process info connection
func init() {
	iobindTool = &IOBindTool{version: version, buildDate: buildDate}
	iobind.New(iobind.WithIOBindCmd("bin/iobind"))
}

// Version number string
func Version() string {
	return iobindTool.version
}

// Build Date string
func BuildDate() string {
	return iobindTool.buildDate
}

func isVerbose() bool {
	return options.Verbose
}

func vPrintf(format string, a ...interface{}) {
	if isVerbose() {
		fmt.Printf(format, a...)
	}
}

func main() {

	fmt.Printf("\n===== IOBind Version: %s, Build Date: %s\n", Version(), BuildDate())

	hlp.SetupSignals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)

	more, err := parser.Parse()
	if err != nil {
		fmt.Printf("*** invalid arguments %v\n", err)
		os.Exit(1)
	}

	if err := iobind.Update(); err != nil {
		fmt.Printf("*** iobind update failed: %v\n", err)
		os.Exit(1)
	}

	if options.BindDev && options.UnbindDev {
		fmt.Printf("\n*** cannot bind and unbind a device at the same time\n")
		os.Exit(1)
	}

	if options.ShowVersion {
		os.Exit(0)
	} else if options.BindDev {
		if len(more) == 0 {
			fmt.Printf("\n*** no device specified for binding\n")
			os.Exit(1)
		}
		list := append([]string{}, more...)
		vPrintf("Binding device(s): %v\n", list)
		if err := iobind.BindPorts(list); err != nil {
			fmt.Printf("*** bind device failed: %v\n", err)
			os.Exit(1)
		}
	} else if options.UnbindDev {
		if len(more) == 0 {
			fmt.Printf("\n*** no device specified for unbinding\n")
			os.Exit(1)
		}
		list := append([]string{}, more...)
		vPrintf("Unbinding device(s): %v\n", list)
		if err := iobind.UnbindPorts(list); err != nil {
			fmt.Printf("*** unbind device failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		netList := iobind.PciNetList()

		fmt.Printf("\nIOBind number of Network Devices (%d)\n", len(netList))
		fmt.Printf("  %-13s %-12s %-12s %s\n", "PCI ID", "Driver", "Module", "Device")

		for _, net := range netList {
			fmt.Printf("  %-13s %-12s %-12s %s\n",
				net.Slot, net.Driver, net.Module, net.Device)
		}
	}

	os.Exit(0)
}
