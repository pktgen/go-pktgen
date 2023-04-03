// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022-2024 Intel Corporation

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

// Make sure the order of the constants above is the same as the order of the
// of the structure below.
type configData struct {
	CoreList             *string   `json:"core-list"`
	NumChannels          int       `json:"num-channels"`
	NumRanks             int       `json:"num-ranks"`
	MemorySize           uint64    `json:"memory-size"`
	InMemory             bool      `json:"in-memory"`
	PortList             []*string `json:"port-list"`
	PortMapping          []*string `json:"port-mapping"`
	FilePrefix           string    `json:"file-prefix"`
	DebugTTY             int       `json:"debug-tty"`
}

type System struct {
	cBytes []byte
	cd     configData
}

func New() *System {

	return &System{
		cBytes: []byte("{}"),
		cd: configData{
			CoreList:             nil,
			NumChannels:          0,
			NumRanks:             0,
			MemorySize:           0, // Size of memory is in MBytes.
			InMemory:             false,
			PortList:             nil,
			PortMapping:          nil,
			FilePrefix:           "",
            DebugTTY:             -1, // -1 means no debug tty used for logging.
		},
	}
}

func (cd *configData) validateConfig() error {

	if cd.CoreList == nil {
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

	if cs.cd.CoreList == nil {
		return ""
	}
	return *cs.cd.CoreList
}

func (cs *System) NumChannels() int {

	return cs.cd.NumChannels
}

func (cs *System) NumRanks() int {

	return cs.cd.NumRanks
}

func (cs *System) MemorySize() uint64 {

	return cs.cd.MemorySize
}

func (cs *System) InMemory() bool {

	return cs.cd.InMemory
}

func (cs *System) PortList() []*string {

	return cs.cd.PortList
}

func (cs *System) PortCount() int {
	return len(cs.cd.PortList)
}

func (cs *System) PortMapping() []*string {

	return cs.cd.PortMapping
}

func (cs *System) FilePrefix() string {

	return cs.cd.FilePrefix
}

func (cs *System) DebugTTY() int {

	return cs.cd.DebugTTY
}

func (cs *System) SetDebugTTY(ptty int) {

	cs.cd.DebugTTY = ptty
}

func (cs *System) MakeArgs() ([]string, error) {

	argv := []string{"dpdk"}

	if cs.CoreList() == "" {
		return nil, fmt.Errorf("core-list option is required")
	} else {
		argv = append(argv, "-l", cs.CoreList())
	}
	if chnls := cs.NumChannels(); chnls > 0 {
		argv = append(argv, "-n", strconv.Itoa(chnls))
	} else {
		return nil, fmt.Errorf("num-channels option is required")
    }
	if ranks := cs.NumRanks(); ranks > 0 {
		argv = append(argv, "-r", strconv.Itoa(ranks))
	}
	if mem := cs.MemorySize(); mem > 0 {
		argv = append(argv, "-m", strconv.FormatUint(mem, 10))
	}
	if cs.InMemory() {
		argv = append(argv, "--in-memory")
	}
	if cs.FilePrefix() != "" {
		argv = append(argv, "--file-prefix", cs.FilePrefix())
	}
	if len(cs.PortList()) > 0 {
		for _, port := range cs.PortList() {
			argv = append(argv, "-a", *port)
		}
    } else {
		return nil, fmt.Errorf("port-list option is required")
	}

	if len(argv) == 0 {
		return nil, fmt.Errorf("no command line arguments specified")
	}

	return argv, nil
}
