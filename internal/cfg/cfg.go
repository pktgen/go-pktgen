// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2024 Intel Corporation

package cfg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/tidwall/jsonc"
)

/*
Go-Pktgen (gPktgen) is a software packet generator for Intel Corp NICs.

This configuration file (pktgen.jsonc) contains various settings for starting
and configuring DPDK started by Go-Pktgen. Not all of the DPDK options are supported
in the Go-Pktgen configuration file.

The JSONC format allows comments and supports basic JSON syntax.
The JSONC format is converted to JSON format before parsing the structured information.

Most of the JSONC data is simple to parse and understood, except for some complex data types.
  Global Options:
  	"debug-tty"    - field supports a single integer value or tty number from /dev/pts/X, e.g., "3".

  DPDK Options:
    "core-list"    - field supports ranges and single lcore numbers, e.g., "0-3,5-7,9,11".
    "num-channels" - number of memory channels, field supports a positive integer value, e.g., default 0.
    "num-ranks"    - number of memory ranks, field supports a positive integer value, e.g., default 0.
    "memory-size"  - field supports a positive integer value in MBytes, e.g., default 256.
    "in-memory"    - field supports a boolean value to have DPDK not use file base memory, e.g.,  default "false".
    "file-prefix"  - field supports a string prefix value for the file based memory names, e.g., "pktgen_" or "rte_map_".
    "pci-list"     - field supports comma-separated list of port PCIe addresses.

  Pktgen Options:
    The "port-mapping" - field supports comma-separated list of port mappings, e.g., "1.0,[2:3].1,[2-3:4-5].2".
	                     format: lcore.port or lcore-list.port or [Rx-lcores:Tx-lcores].port"
    The "promiscuous"  - field supports a boolean value, e.g., default "true".
*/

type OptionsDPDK struct {
	CoreList    *string  `json:"core-list"`    // Comma-separated list of core numbers and ranges.
	NumChannels int      `json:"num-channels"` // Number of memory channels in the system.
	NumRanks    int      `json:"num-ranks"`    // Number of memory ranks in the system.
	MemorySize  uint64   `json:"memory-size"`  // Size of memory in MBytes.
	InMemory    bool     `json:"in-memory"`    // Whether to use in-memory mode.
	FilePrefix  string   `json:"file-prefix"`  // Prefix for the file names.
	PciList     []string `json:"pci-list"`     // Comma-separated list of port PCIe addresses.
}

type OptionsPktgen struct {
	PortMapping []string `json:"port-mapping"` // Comma-separated list of port mappings.
	Promiscuous bool     `json:"promiscuous"`  // Whether to enable promiscuous mode. Default: true.
}

// Make sure the order of the constants above is the same as the order of the
// of the structure below.
type configData struct {
	DebugTTY int           `json:"debug-tty"` // Debug tty for logging.
	DPDK     OptionsDPDK   `json:"dpdk"`      // DPDK options.
	Pktgen   OptionsPktgen `json:"pktgen"`    // Pktgen options.
}

type System struct {
	cBytes []byte     // The JSONC configuration data converted to JSON format and bytes.
	cd     configData // The configuration information.
}

func New() *System {

	return &System{
		cBytes: []byte("{}"),
		cd: configData{
			DebugTTY: -1, // -1 means no debug tty used for logging.
			DPDK: OptionsDPDK{
				CoreList:    nil,
				NumChannels: 0,
				NumRanks:    0,
				MemorySize:  0, // Size of memory is in MBytes.
				InMemory:    false,
				FilePrefix:  "",
				PciList:     nil,
			},
			Pktgen: OptionsPktgen{
				PortMapping: nil,
				Promiscuous: false,
			},
		},
	}
}

func (cd *configData) validateConfig() error {

	if cd.DPDK.CoreList == nil || *cd.DPDK.CoreList == "" {
		return fmt.Errorf("core-list is empty")
	}
	return nil
}

func (cs *System) openText() error {

	text := jsonc.ToJSON(bytes.TrimSpace(cs.cBytes))

	if len(text) == 0 {
		return fmt.Errorf("empty json text string")
	}

	// test for JSON string, which must start with a '{'
	if text[0] != '{' {
		return fmt.Errorf("string does not appear to be a valid JSON text missing starting '{'")
	}

	// Unmarshal json text into the Config structure
	if err := json.Unmarshal(text, &cs.cd); err != nil {
		return err
	}
	return cs.cd.validateConfig()
}

// readFile by passing in a filename or path to a JSON-C or JSON configuration
func (cs *System) readFile(s string) error {
	b, err := os.ReadFile(s)
	if err != nil {
		return err
	}
	cs.cBytes = b
	return nil
}

func (cs *System) Open(s string) error {

	if len(s) == 0 {
		s = "{}"
	}
	if err := cs.readFile(s); err != nil {
		cs.cBytes = []byte(s)
	}
	return cs.openText()
}

func (cs *System) String() string {

	if data, err := json.MarshalIndent(&cs.cd, "", "  "); err != nil {
		return fmt.Sprintf("error marshalling JSON: %v", err)
	} else {
		return string(data)
	}
}

// *** DPDK Options ***

func (cs *System) CoreList() string {

	if cs.cd.DPDK.CoreList == nil {
		return ""
	}
	return *cs.cd.DPDK.CoreList
}

func (cs *System) NumChannels() int {

	return cs.cd.DPDK.NumChannels
}

func (cs *System) NumRanks() int {

	return cs.cd.DPDK.NumRanks
}

func (cs *System) MemorySize() uint64 {

	return cs.cd.DPDK.MemorySize
}

func (cs *System) InMemory() bool {

	return cs.cd.DPDK.InMemory
}

func (cs *System) FilePrefix() string {

	return cs.cd.DPDK.FilePrefix
}

func (cs *System) PciList() []string {

	return cs.cd.DPDK.PciList
}

func (cs *System) PortCount() int {
	return len(cs.cd.DPDK.PciList)
}

// *** Pktgen Options ***

func (cs *System) PortMapping() []string {

	return cs.cd.Pktgen.PortMapping
}

func (cs *System) PortMappingCount() int {

	return len(cs.cd.Pktgen.PortMapping)
}

func (cs *System) DebugTTY() int {

	return cs.cd.DebugTTY
}

func (cs *System) SetDebugTTY(ptty int) {

	cs.cd.DebugTTY = ptty
}

func (cs *System) Promiscuous() bool {
	return cs.cd.Pktgen.Promiscuous
}

// Return the command-line arguments for starting DPDK.
func (cs *System) GetArgsDPDK() ([]string, error) {

	args := []string{}

	if cs.CoreList() == "" {
		return nil, fmt.Errorf("core-list option is required")
	} else {
		args = append(args, "-l", cs.CoreList())
	}
	if chnls := cs.NumChannels(); chnls > 0 {
		args = append(args, "-n", strconv.Itoa(chnls))
	} else {
		return nil, fmt.Errorf("num-channels option is required")
	}
	if ranks := cs.NumRanks(); ranks > 0 {
		args = append(args, "-r", strconv.Itoa(ranks))
	}
	if mem := cs.MemorySize(); mem > 0 {
		args = append(args, "-m", strconv.FormatUint(mem, 10))
	}
	if cs.InMemory() {
		args = append(args, "--in-memory")
	}
	if cs.FilePrefix() != "" {
		args = append(args, "--file-prefix", cs.FilePrefix())
	}

	if len(cs.PciList()) > 0 {
		for _, port := range cs.PciList() {
			args = append(args, "-a", port)
		}
	} else {
		return nil, fmt.Errorf("pci-list option is required")
	}

	return args, nil
}
