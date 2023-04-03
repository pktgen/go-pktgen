// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"syscall"

	flags "github.com/jessevdk/go-flags"

	gpkt "github.com/pktgen/go-pktgen/internal/gopktgen"
	hlp "github.com/pktgen/go-pktgen/internal/helpers"
)

var (
	pktgenApp *PktgenApp = &PktgenApp{} // Main application object singleton

	// The following two variables are filled in by the build script
	version   string // Version of Pktgen, filled in build script
	buildDate string // Build date of Pktgen, filled in build script
)

func (pa *PktgenApp) parseOptions() error {

	parser  := flags.NewParser(&pa.options, flags.Default)

    // Parse the command line options
    if args, err := parser.Parse(); err != nil {
    	return err
    } else {
		if len(args) > 0 {
			return fmt.Errorf("unknown command-line argument(s): %+v\n", args)
		}
	}

    // If the -V or --version option is given, print the version and exit
    if pa.options.ShowVersion {
        fmt.Printf("==== %s %s (Build Date: %s)\n", ApplicationTitle, version, buildDate)
		os.Exit(0)
    }

	if pa.options.Ptty > 0 {
		pa.pCfg.PseudoTTY = pa.options.Ptty
	}

	// If the -c or --config-file option is given, parse the configuration file
	if pCfg, err := gpkt.ParseConfigFile(pa.options.ConfigFile); err != nil {
		return err
    } else {
		pa.pCfg = pCfg
	}

	// Command line option overrides the configuration file option for PseudoTTY
	if pa.options.Ptty > 0 {
		pa.pCfg.PseudoTTY = pa.options.Ptty
	}

	if err := pktgenApp.StartLogging(); err != nil {
		return err
	}

	return nil
}

func main() {

	// Setup some values for the version and build information in helpers package
	hlp.VersionStr = version
	hlp.BuildDateStr = buildDate
	hlp.ApplicationStr = ApplicationTitle
	hlp.CopyrightStr = Copyright

    // Set up signal handlers for signal termination events
    hlp.SetupSignals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)

	// Parse the command line options
	if err := pktgenApp.parseOptions(); err != nil {
		log.Fatalf("*** Error: %v\n", err)
    }

	var f *os.File

	// If the -e or --prof-enable option is given, start profiling
	if pktgenApp.options.ProfEnable {
		if f, err := os.Create("cpu-profile"); err != nil {
			log.Fatalf("failed to create CPU profile: %v\n", err)
        } else {
			pprof.StartCPUProfile(f)
		}
	}
	defer func() {
		pprof.StopCPUProfile()
		f.Close() // Close the output file after profiling is done
	}()

	if err := pktgenApp.Start(); err != nil {
		log.Fatalf("failed to start Go-Pktgen: %s\n", err)
	}
}
