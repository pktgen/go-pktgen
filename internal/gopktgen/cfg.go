// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package gopktgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"

	gpc "github.com/pktgen/go-pktgen/internal/gpcommon"
	tlog "github.com/pktgen/go-pktgen/internal/tlog"
	"github.com/tidwall/jsonc"
)

type System struct {
	config string            // Path to the configuration file.
	cBytes []byte            // The JSONC configuration data converted to JSON format and bytes.
	pCfg   gpc.PktgenConfig // The configuration information.
}

type CfgOption func(*System)

func WithConfig(config string) CfgOption {
	return func(s *System) {
		// Open the configuration file or JSON string
		if config[0] == '{' {
			s.cBytes = []byte(config)
		} else {
			if err := s.readFile(config); err != nil {
				fmt.Printf("Error reading configuration file: %v\n", err)
				os.Exit(1)
			}
		}
	}
}

func NewConfig(options ...CfgOption) (*System, error) {

	s := &System{
		cBytes: []byte("{}"),
		pCfg: gpc.PktgenConfig{
			LogTTY: -1, // -1 means no debug tty used for logging.
			DPDK: gpc.OptionsDPDK{
				NumChannels: gpc.DefaultNumChannels,
				NumRanks:    gpc.DefaultNumRanks,
				MemorySize:  gpc.DefaultMemorySize, // Size of memory is in MBytes.
				InMemory:    gpc.DefaultInMemory,
				FilePrefix:  gpc.DefaultFilePrefix,
			},
			Pktgen: gpc.OptionsPktgen{
				Cards:       make([]string, 0),
				Mappings:    make([]gpc.Mapping, 0),
				Promiscuous: gpc.DefaultPromiscuousMode,
			},
		},
	}

	// Process the option function calls
	for _, option := range options {
		option(s)
	}

	return s, s.openText()
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
	if err := json.Unmarshal(text, &cs.pCfg); err != nil {
		return err
	}

	tlog.Printf("Parsed configuration:\n%s\n", text)
	gpc.MarshalIndent(&cs.pCfg)

	for core, mapping := range cs.pCfg.Pktgen.Mappings {
		if mapping.Core >= gpc.MaxLogicalCores {
			return fmt.Errorf("invalid core ID: %d", mapping.Core)
		}
		mapping.Core = gpc.CoreID(core)
		if mapping.Mode.Value() == gpc.MainMode {
			mapping.Port = gpc.MaxEtherPorts
		}
	}

	return cs.pCfg.ValidateConfig()
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

func (cs *System) CoreList() string {

	corelist := ""

	cores := make(sort.IntSlice, 0)
	for _, m := range cs.pCfg.Pktgen.Mappings {
		cores = append(cores, int(m.Core))
	}
	cores.Sort()

	for _, core := range cores {
		corelist += fmt.Sprintf("%d,", core)
	}
	if len(corelist) > 0 {
		corelist = corelist[:len(corelist)-1] // remove trailing comma
	}

	return corelist
}

func (cs *System) NumChannels() uint16 {

	return cs.pCfg.DPDK.NumChannels
}

func (cs *System) NumRanks() uint16 {

	return cs.pCfg.DPDK.NumRanks
}

func (cs *System) MemorySize() uint64 {

	return cs.pCfg.DPDK.MemorySize // Size of memory is in Bytes
}

func (cs *System) InMemory() bool {

	return cs.pCfg.DPDK.InMemory
}

func (cs *System) FilePrefix() string {

	return cs.pCfg.DPDK.FilePrefix
}

func (cs *System) PciList() []string {

	return cs.pCfg.Pktgen.Cards
}

func (cs *System) PortCount() int {
	return len(cs.pCfg.Pktgen.Cards)
}

func (cs *System) Mappings() []*gpc.Mapping {

	mappings := make(sort.IntSlice, 0)
	for core := range cs.pCfg.Pktgen.Mappings {
		mappings = append(mappings, int(core))
	}
	mappings.Sort()

	maps := make([]*gpc.Mapping, 0)
	for _, core := range mappings {
		maps = append(maps, &cs.pCfg.Pktgen.Mappings[gpc.CoreID(core)])
	}

	return maps
}

func (cs *System) MappingCount() int {

	return len(cs.pCfg.Pktgen.Mappings)
}

func (cs *System) LogTTY() int {

	return cs.pCfg.LogTTY
}

func (cs *System) SetLogTTY(ptty int) {

	cs.pCfg.LogTTY = ptty
}

func (cs *System) Promiscuous() bool {
	return cs.pCfg.Pktgen.Promiscuous
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
		args = append(args, "-n", strconv.Itoa(int(chnls)))
	} else {
		return nil, fmt.Errorf("num-channels option is required")
	}
	if ranks := cs.NumRanks(); ranks > 0 {
		args = append(args, "-r", strconv.Itoa(int(ranks)))
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
		for _, pci := range cs.PciList() {
			args = append(args, "-a", pci)
		}
	} else {
		return nil, fmt.Errorf("pci-list option is required")
	}

	return args, nil
}
