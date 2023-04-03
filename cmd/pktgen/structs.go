// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package main

import (
	gpkt "github.com/pktgen/go-pktgen/internal/gopktgen"
	"github.com/pktgen/go-pktgen/pkgs/kview"
)

const (
	ApplicationTitle string = "Go-Pktgen Traffic Generator"
	Copyright        string = "Copyright © 2023-2025 Intel Corporation"

	apiLibName  string = "libgpkt_gapi.so"
	hmapLibName string = "libgpkt_hmap.so"
	modeLibName string = "libgpkt_modes.so"
	tlogLibName string = "libgpkt_tlog.so"
)

// Options command line options
type Options struct {
	ConfigFile  string `short:"c" long:"config-file" description:"JSON configuration file or string"`
	Ptty        int    `short:"p" long:"ptty" description:"Enable pseudo-TTY mode (for debugging)" default:"0"`
	ProfEnable  bool   `short:"e" long:"prof-enable" description:"Enable profiling"`
	ShowVersion bool   `short:"V" long:"version" description:"Print out version and exit"`
}

type PktgenApp struct {
	appView *kview.Application // Main application structure
	kPanels *kview.Panels      // Panels for Go-Pktgen
	pCfg    *gpkt.PktgenConfig    // Parsed configuration structure
	gPkt    *gpkt.GoPktgen     // Go-Pktgen API structure
	options Options            // Command line options
}
