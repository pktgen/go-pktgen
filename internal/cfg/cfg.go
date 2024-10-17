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
in the Go-Pktgen configuration file and the supported options are listed below using
a '>' prefix.

The JSONC
  UsePortList [<PCI bus:device.function>,...] Use a specific PCI device
*/

type OptionsDPDK struct {
	CoreList    *string `json:"core-list"`
	NumChannels int     `json:"num-channels"`
	NumRanks    int     `json:"num-ranks"`
	MemorySize  uint64  `json:"memory-size"`
	InMemory    bool    `json:"in-memory"`
	FilePrefix  string  `json:"file-prefix"`
}

type OptionsPktgen struct {
	PortList    []string `json:"port-list"`
	PortMapping []string `json:"port-mapping"`
	Promiscuous bool     `json:"promiscuous"`
}

// Make sure the order of the constants above is the same as the order of the
// of the structure below.
type configData struct {
	DebugTTY int           `json:"debug-tty"`
	DPDK     OptionsDPDK   `json:"dpdk"`
	Pktgen   OptionsPktgen `json:"pktgen"`
}

type System struct {
	cBytes   []byte
	argvList []string
	cd       configData
}

func New() *System {

	return &System{
		cBytes:   []byte("{}"),
		argvList: []string{},
		cd: configData{
			DebugTTY: -1, // -1 means no debug tty used for logging.
			DPDK: OptionsDPDK{
				CoreList:    nil,
				NumChannels: 0,
				NumRanks:    0,
				MemorySize:  0, // Size of memory is in MBytes.
				InMemory:    false,
				FilePrefix:  "",
			},
			Pktgen: OptionsPktgen{
				PortList:    nil,
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

func (cs *System) PortList() []string {

	return cs.cd.Pktgen.PortList
}

func (cs *System) PortCount() int {
	return len(cs.cd.Pktgen.PortList)
}

func (cs *System) PortMapping() []string {

	return cs.cd.Pktgen.PortMapping
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

func (cs *System) processDPDK() error {

	if cs.CoreList() == "" {
		return fmt.Errorf("core-list option is required")
	} else {
		cs.argvList = append(cs.argvList, "-l", cs.CoreList())
	}
	if chnls := cs.NumChannels(); chnls > 0 {
		cs.argvList = append(cs.argvList, "-n", strconv.Itoa(chnls))
	} else {
		return fmt.Errorf("num-channels option is required")
	}
	if ranks := cs.NumRanks(); ranks > 0 {
		cs.argvList = append(cs.argvList, "-r", strconv.Itoa(ranks))
	}
	if mem := cs.MemorySize(); mem > 0 {
		cs.argvList = append(cs.argvList, "-m", strconv.FormatUint(mem, 10))
	}
	if cs.InMemory() {
		cs.argvList = append(cs.argvList, "--in-memory")
	}
	if cs.FilePrefix() != "" {
		cs.argvList = append(cs.argvList, "--file-prefix", cs.FilePrefix())
	}

	if len(cs.PortList()) > 0 {
		for _, port := range cs.PortList() {
			cs.argvList = append(cs.argvList, "-a", port)
		}
	} else {
		return fmt.Errorf("port-list option is required")
	}

	return nil
}

func (cs *System) processPktgen() error {

	// Adding the '--' separator to indicate the end of DPDK options
	cs.argvList = append(cs.argvList, "--")

	if cs.Promiscuous() {
		cs.argvList = append(cs.argvList, "-P")
	}

	if len(cs.PortMapping()) > 0 {
		for _, m := range cs.PortMapping() {
			cs.argvList = append(cs.argvList, "-m", m)
		}
	} else {
		return fmt.Errorf("port-mapping option is required")
	}

	return nil
}

func (cs *System) CreateArgs() ([]string, error) {

	cs.argvList = []string{} // Reset the command line arguments

	// Start DPDK options adding the '--' separator to indicate the end of Pktgen options
	if err := cs.processDPDK(); err != nil {
		return nil, err
	}
	// End of DPDK options, start Pktgen options
	if err := cs.processPktgen(); err != nil {
		return nil, err
	}

	if len(cs.argvList) == 0 {
		return nil, fmt.Errorf("no command line arguments specified")
	}

	return cs.argvList, nil
}
