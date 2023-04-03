// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package main

import (
	"log"
	"os"
	"runtime/pprof"
	"syscall"

	flags "github.com/jessevdk/go-flags"

	hlp "github.com/pktgen/go-pktgen/internal/helpers"
)

var (
	pktgenApp *PktgenApp // Main application object singleton

	// The following two variables are filled in by the build script
	version   string // Version of Pktgen, filled in build script
	buildDate string // Build date of Pktgen, filled in build script
)

func main() {

	pg := PktgenApp{
		options: Options{},
	}

	// Setup some values for the version and build information in helpers package
	hlp.VersionStr = version
	hlp.BuildDateStr = buildDate
	hlp.ApplicationStr = ApplicationTitle
	hlp.CopyrightStr = Copyright

    // Set up signal handlers for termination
    hlp.SetupSignals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)

	parser  := flags.NewParser(&pg.options, flags.Default) // Command line parser object for command line parsing

    // Parse the command line options
    _, err := parser.Parse()
    if err!= nil {
        log.Fatalf("invalid arguments %v\n", err)
    }

    // If the -V or --version option is given, print the version and exit
    if pg.options.ShowVersion {
        os.Exit(0)
    }

	var f *os.File

	// If the -e or --prof-enable option is given, start profiling
	if pg.options.ProfEnable {
		if f, err := os.Create("cpu-profile"); err != nil {
			log.Fatalf("failed to create CPU profile file: %v\n", err)
        } else {
			pprof.StartCPUProfile(f)
		}
	}
	defer func() {
		pprof.StopCPUProfile()
		f.Close() // Close the output file after profiling is done
	}()

	if err := pg.Start(); err != nil {
		log.Fatalf("failed to start Go-Pktgen: %s\n", err)
	}
}
