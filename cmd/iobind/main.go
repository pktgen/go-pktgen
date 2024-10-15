// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	version  string             // Version of tool
}

// Options command line options
type Options struct {
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
}

// Version number string
func (st *IOBindTool) Version() string {
	return st.version
}

func main() {

	fmt.Printf("\n===== IOBind version: %s =====\n", iobindTool.Version())

	setupSignals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)

	if _, err := parser.Parse(); err != nil {
		fmt.Printf("*** invalid arguments %v\n", err)
		os.Exit(1)
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
