// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package main

import (
	"fmt"

	gpkt "github.com/pktgen/go-pktgen/internal/gopktgen"
	gpc "github.com/pktgen/go-pktgen/internal/gpcommon"
	"github.com/pktgen/go-pktgen/internal/iobind"
	"github.com/pktgen/go-pktgen/internal/tlog"

	hlp "github.com/pktgen/go-pktgen/internal/helpers"
)

func (pg *PktgenApp) setupDPDKArgs() error {
	// Create the DPDK arguments
	argv, err := pg.config.GetArgsDPDK()
	if err != nil {
		return fmt.Errorf("failed to get DPDK arguments: %s", err)
	}
	if len(argv) == 0 {
		return fmt.Errorf("DPDK arguments are empty")
	}

	// Set the C-compatible array of strings
	pg.gPkt.AddArgv(gpc.DPDKProgramName) // Add the program name
	for _, s := range argv {
		if pg.gPkt.AddArgv(s) < 0 {
			fmt.Errorf("failed to add DPDK argument: %s", s)
		}
	}
	return nil
}

func (pg *PktgenApp) StartLogging() error {

	// Initialize the tty logger if enabled.
	if pg.config.LogTTY() > 0 {
		if err := tlog.Open(tlog.WithLogID(pg.config.LogTTY())); err != nil {
			return fmt.Errorf("failed to open tty logger: %s", err)
		}
		tlog.Printf("Starting Go-Pktgen with Ptty: /dev/pts/%d\n", pg.config.LogTTY())
	}

	// Command line option overrides the configuration file
	if pg.options.Ptty > 0 {
		pg.config.SetLogTTY(pg.options.Ptty)
	}

	str := hlp.CommandInfo(false)
	tlog.Printf("===== %s\n", str)
	tlog.Printf("===== Build Date: %s\n", buildDate)

	return nil
}

func (pg *PktgenApp) LoadLibraries() error {

	// Keep the order up to date with the library names in libNames.
	libNames := []string{tlogLibName, hmapLibName, modeLibName, apiLibName} // Order matters here!

	// Register the libraries and APIs to the main application
	if g, err := gpkt.New(gpkt.WithLibNames(libNames), gpkt.WithLogPath(pg.config.LogTTY())); err != nil {
		return fmt.Errorf("invalid arguments %v", err)
	} else {
		pg.gPkt = g
	}
	return nil
}

func (pg *PktgenApp) Start() error {
	// Set up the application object and fill in the version and build date
	pktgenApp = pg

	str := hlp.CommandInfo(false)
	fmt.Printf("===== %s\n", str)
	fmt.Printf("===== Build Date: %s\n", buildDate)

	// Parse the JSON configuration file.
	if cfg, err := gpkt.NewConfig(gpkt.WithConfig(pg.options.ConfigData)); err != nil {
		return fmt.Errorf("load configuration file %s failed: %s", pg.options.ConfigData, err)
	} else {
		pg.config = cfg
	}

	// Set up the application object and fill in the version and build date
	if err := pg.StartLogging(); err != nil {
		return fmt.Errorf("load configuration failed: %s", err)
	}

	// Load the libraries and APIs to the main application.
	if err := pg.LoadLibraries(); err != nil {
		return fmt.Errorf("load libraries failed: %s", err)
	}

	// Add the PCI devices to the network interface vfio-pci to be used by DPDK.
	pg.gPkt.AddPorts(pg.config.PciList()...)

	if err := pg.setupDPDKArgs(); err != nil {
		return fmt.Errorf("failed to setup DPDK arguments: %s", err)
	}

	// Add the mapping strings to gopktgen and then process the mapping strings.
	pg.gPkt.AddMapping(pg.config.Mappings()...)
	if err := pg.gPkt.ProcessMaps(); err != nil {
		return err
	}

	// Set up the L2P configuration.
	if err := pg.gPkt.L2pConfigSet(); err != nil {
		return err
	}

	// Attempt to bind devices to the network interface vfio-pci to be used by DPDK.
	if iob := iobind.New(iobind.WithIOBindCmd("bin/iobind")); iob == nil {
		return fmt.Errorf("failed to initialize IOBind")
	}

	// Bind the ports to the network interface vfio-pci to be used by DPDK.
	if err := iobind.IOBindPorts(pg.config.PciList()); err != nil {
		return err
	}

	// Initialize the main application and panels for Go-Pktgen
	if err := pg.initPanels(); err != nil {
		return fmt.Errorf("failed to initialize Pktgen panels: %s", err)
	} else {
		if err := pg.gPkt.Start(); err != nil {
			return fmt.Errorf("failed to start GoPktgen: %s", err)
		}

		defer func() {
			pg.gPkt.Stop()
			pg.gPkt.Destroy()
		}()

		if err := pg.appView.Run(); err != nil {
			return fmt.Errorf("failed to run Pktgen application: %s", err)
		}
	}
	return nil
}
