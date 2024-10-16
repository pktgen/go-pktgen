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
	BindDev     string `short:"b" long:"bind" description:"Bind device to a driver"`
	UnBindDev   string `short:"u" long:"unbind" description:"Unbind device from a driver"`
	ShowVersion bool   `short:"V" long:"version" description:"Print out version and exit"`
	Verbose     bool   `short:"v" long:"verbose" description:"Output verbose messages"`
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
func (st *IOBindTool) Version() string {
	return st.version
}

func main() {

	fmt.Printf("\n===== IOBind version: %s =====\n", iobindTool.Version())

	setupSignals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)

	devices, err := parser.Parse()
	if err != nil {
		fmt.Printf("*** invalid arguments %v\n", err)
		os.Exit(1)
	}

	iobindTool.db.Start()

	if options.ShowVersion {
        os.Exit(0)
    } else if options.BindDev != "" {
		if err := iobindTool.db.BindPorts(devices); err!= nil {
            fmt.Printf("*** bind device failed: %v\n", err)
            os.Exit(1)
        }
	} else if options.UnBindDev != "" {
		if err := iobindTool.db.UnbindPorts(devices); err!= nil {
            fmt.Printf("*** unbind device failed: %v\n", err)
            os.Exit(1)
        }
	} else {
		netList := iobindTool.db.NetList()

		fmt.Printf("Network Devices (%d)\n", len(netList))
		fmt.Printf("  %-13s %-12s %-12s %s\n", "PCI ID", "Driver", "Module", "Device")

		for _, net := range netList {
			fmt.Printf("  %-13s %-12s %-12s %s\n",
				net.Slot, net.Driver, net.Module, net.Device)
		}
	}
}

func setupSignals(signals ...os.Signal) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, signals...)
	go func() {
		<-sigs

		os.Exit(1)
	}()
}
