// SPDX-License-Identifier: BSD-3-Clause
// Copyright(c) 2022-2024 Intel Corporation

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/BurntSushi/toml"
	flags "github.com/jessevdk/go-flags"
	"github.com/pktgen/go-pktgen/internal/fserde"
)

const (
	toolVersion = "24.09.0"
)

type frame struct {
	data string
}

// SerdeTool to convert a text file to a packet file.
type SerdeTool struct {
	version  string             // Version of tool
	serde    *fserde.FrameSerde // pointer to fserde.FrameSerde structure
	tomlData struct {
		OutputPcapFile string  `toml:"pcap-output-file"`
		Packets        []frame `toml:"Packets"`
		Defaults       []frame `toml:"Defaults"`
	}
}

// Options command line options
type Options struct {
	FileToml    string `short:"t" long:"file-toml" description:"TOML file containing frame strings" value-name:"<file>"`
	PcapFile    string `short:"p" long:"pcap-file" description:"PCAP file name" value-name:"<file>"`
	ShowVersion bool   `short:"V" long:"version" description:"Print out version and exit"`
	Verbose     bool   `short:"v" long:"verbose" description:"Output verbose messages"`
}

// Global to the main package for the tool
var serdeTool *SerdeTool
var options Options
var parser = flags.NewParser(&options, flags.Default)

// Setup the tool's global information and startup the process info connection
func init() {
	serdeTool = &SerdeTool{}
	serdeTool.version = toolVersion
	if serde, err := fserde.Create("SerdeTool", &fserde.FrameSerdeConfig{}); err != nil {
		panic(err)
	} else {
		serdeTool.serde = serde
	}
}

// Version number string
func (st *SerdeTool) Version() string {
	return st.version
}

func (a *frame) UnmarshalText(text []byte) error {

	a.data = strings.ReplaceAll(strings.TrimSpace(string(text)), "\n", "")
	return nil
}

func main() {

	fmt.Printf("\n===== Frame Serde version: %s =====\n", serdeTool.Version())

	setupSignals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)

	if _, err := parser.Parse(); err != nil {
		fmt.Printf("*** invalid arguments %v\n", err)
		os.Exit(1)
	}

	if len(options.FileToml) > 0 {
		if data, err := os.ReadFile(options.FileToml); err != nil {
			fmt.Printf("load frame file failed: %s\n", err)
			os.Exit(1)
		} else {
			if meta, err := toml.Decode(string(data), &serdeTool.tomlData); err != nil {
				fmt.Printf("decoding toml file failed: %s\n", err)
				os.Exit(1)
			} else {
				if len(meta.Undecoded()) > 0 {
					fmt.Printf("*** undecoded items %s\n", meta.Undecoded())
					os.Exit(1)
				}
				var packets []string
				for _, p := range serdeTool.tomlData.Defaults {
					packets = append(packets, p.data)
				}
				defs := &fserde.FrameSerdeConfig{Defaults: packets}
				if fg, err := fserde.Create("Convert", defs); err != nil {
					fmt.Printf("*** create failed %v\n", err)
					os.Exit(1)
				} else {
					packets = nil
					for _, p := range serdeTool.tomlData.Packets {
						packets = append(packets, p.data)
					}
					if err := fg.StringsToBinary(packets); err != nil {
						fmt.Printf("*** string to binary failed %v\n", err)
						os.Exit(1)
					}
					if len(options.PcapFile) > 0 {
						if err := fg.WritePCAP(options.PcapFile, fserde.NormalFrameType); err != nil {
							fmt.Printf("*** write pcap failed %v\n", err)
							os.Exit(1)
						}
					} else {
						fmt.Printf("\n")
						for _, pkt := range fg.GetFrames(fserde.NormalFrameType) {
							s := strings.Split(fmt.Sprintf("%v", pkt), "/")
							for i, v := range s {
								if i == 0 {
									fmt.Printf("%v/\n", v)
								} else {
									fmt.Printf("    %s", v)
									if i == len(s)-1 {
										fmt.Printf("\n")
									} else {
										fmt.Printf("/\n")
									}
								}
							}
						}
					}
				}
			}
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
