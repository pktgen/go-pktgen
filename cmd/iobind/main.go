// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pktgen/go-pktgen/internal/devbind"

	flags "github.com/jessevdk/go-flags"
)

const (
	toolVersion = "24.09.0"
)

type frame struct {
	data string
}

// IOBindTool to convert a text file to a packet file.
type IOBindTool struct {
	version string           // Version of tool
	db      *devbind.DevBind // Device binding object
}

// Options command line options
type Options struct {
	BindDev     []string `short:"b" long:"bind" description:"Bind device to a driver"`
	UnBindDev   []string `short:"u" long:"unbind" description:"Unbind device from a driver"`
	ShowVersion bool     `short:"v" long:"version" description:"Print out version and exit"`
	Verbose     bool     `short:"V" long:"verbose" description:"Enable verbose output"`
}

// Global to the main package for the tool
var iobindTool *IOBindTool
var options Options
var parser = flags.NewParser(&options, flags.Default)

// Setup the tool's global information and startup the process info connection
func init() {
	iobindTool = &IOBindTool{version: toolVersion}
	iobindTool.db = devbind.New()
}

// Version number string
func Version() string {
	return iobindTool.version
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

	setupSignals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)

	more, err := parser.Parse()
	if err != nil {
		fmt.Printf("*** invalid arguments %v\n", err)
		os.Exit(1)
	}

	iobindTool.db.Start()

	if options.ShowVersion {
		fmt.Printf("\nIOBind Version: %s\n", Version())
		os.Exit(0)
	} else if len(options.BindDev) > 0 {
		if len(more) > 0 {
			options.BindDev = append(options.BindDev, more...)
		}
		vPrintf("Binding device(s): %v\n", options.BindDev)
		if err := iobindTool.db.BindPorts(options.BindDev); err != nil {
			vPrintf("\nIOBind Version: %s\n", Version())
			fmt.Printf("*** bind device failed: %v\n", err)
			os.Exit(1)
		}
	} else if len(options.UnBindDev) > 0 {
		if len(more) > 0 {
			options.UnBindDev = append(options.UnBindDev, more...)
		}
		vPrintf("Unbinding device(s): %v\n", options.UnBindDev)
		if err := iobindTool.db.UnbindPorts(options.UnBindDev); err != nil {
			vPrintf("\nIOBind Version: %s\n", Version())
			fmt.Printf("*** unbind device failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		netList := iobindTool.db.NetList()

		fmt.Printf("\nIOBind Version: %s Network Devices (%d)\n", Version(), len(netList))
		fmt.Printf("  %-13s %-12s %-12s %s\n", "PCI ID", "Driver", "Module", "Device")

		for _, net := range netList {
			fmt.Printf("  %-13s %-12s %-12s %s\n",
				net.Slot, net.Driver, net.Module, net.Device)
		}
	}

	iobindTool.db.Stop()
}

func setupSignals(signals ...os.Signal) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, signals...)
	go func() {
		<-sigs

		os.Exit(1)
	}()
}
