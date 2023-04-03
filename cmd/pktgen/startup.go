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

func (pa *PktgenApp) setupDPDKArgs(pc *gpkt.PktgenConfig) error {
	// Create the DPDK arguments
	argv, err := pc.GetArgsDPDK()
	if err != nil {
		return fmt.Errorf("failed to get DPDK arguments: %s", err)
	}
	if len(argv) == 0 {
		return fmt.Errorf("DPDK arguments are empty")
	}

	// Set the C-compatible array of strings
	pa.gPkt.AddArgv(gpc.DPDKProgramName) // Add the program name

	// Add the DPDK arguments to the Go-Pktgen argument list
	for _, s := range argv {
		if pa.gPkt.AddArgv(s) < 0 {
			fmt.Errorf("failed to add DPDK argument: %s", s)
		}
	}
	return nil
}

func (pg *PktgenApp) StartLogging() error {

	// Initialize the tty logger if enabled.
	if pg.pCfg.PseudoTTY > 0 {
		if err := tlog.Open(tlog.WithLogID(pg.pCfg.PseudoTTY)); err != nil {
			return fmt.Errorf("failed to open tty logger: %s", err)
		}
	}

	str := hlp.CommandInfo(false)
	tlog.Printf("===== %s\n", str)
	tlog.Printf("===== Build Date: %s\n", buildDate)

	return nil
}

func (pa *PktgenApp) Start() error {

	str := hlp.CommandInfo(false)
	fmt.Printf("===== %s\n", str)
	fmt.Printf("===== Build Date: %s\n", buildDate)

	// Keep the order up to date with the library names in libNames.
	libNames := []string{tlogLibName, hmapLibName, modeLibName, apiLibName} // Order matters here!

	// Register the libraries and APIs to the main application
	g, err := gpkt.New(pa.pCfg, gpkt.WithLogPath(pa.pCfg.PseudoTTY), gpkt.WithLibNames(libNames))
	if err != nil {
		return fmt.Errorf("invalid arguments %v", err)
	} else {
		pa.gPkt = g
	}

	// Add the PCI devices to the network interface vfio-pci to be used by DPDK.
	pa.gPkt.AddPorts(pa.pCfg.PciList()...)

	if err := pa.setupDPDKArgs(pa.pCfg); err != nil {
		return fmt.Errorf("failed to setup DPDK arguments: %s", err)
	}

	// Add the mapping strings to gopktgen and then process the mapping strings.
	pa.gPkt.AddMapping(pa.pCfg.BaseMappings()...)
	if err := pa.gPkt.ProcessMaps(); err != nil {
		return err
	}

	// Set up the L2P configuration.
	if err := pa.gPkt.L2pConfigSet(); err != nil {
		return err
	}

	// Attempt to bind devices to the network interface vfio-pci to be used by DPDK.
	if iob := iobind.New(iobind.WithIOBindCmd("bin/iobind")); iob == nil {
		return fmt.Errorf("failed to initialize IOBind")
	}

	// Bind the ports to the network interface vfio-pci to be used by DPDK.
	if err := iobind.IOBindPorts(pa.pCfg.PciList()); err != nil {
		return err
	}

	// Initialize the main application and panels for Go-Pktgen
	if err := pa.initPanels(); err != nil {
		return fmt.Errorf("failed to initialize Pktgen panels: %s", err)
	} else {
		if err := pa.gPkt.Start(); err != nil {
			return fmt.Errorf("failed to start GoPktgen: %s", err)
		}

		defer func() {
			pa.gPkt.Stop()
			pa.gPkt.Destroy()
		}()

		if err := pa.appView.Run(); err != nil {
			return fmt.Errorf("failed to run Pktgen application: %s", err)
		}
	}
	return nil
}
