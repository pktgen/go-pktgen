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
	"github.com/tidwall/jsonc"
)

func ParseConfigFile(str string) (*PktgenConfig, error) {

	pCfg := &PktgenConfig{}

	text := bytes.TrimSpace([]byte(str))
	if len(text) == 0 {
		fmt.Printf("*** Configuration string is empty\n")
		os.Exit(1)
	}

	// If the string starts with '{' it's a JSON string, otherwise it's a file path
	if text[0] != '{' {
		// assume it's a file path to a JSON or JSONc file
		if data, err := os.ReadFile(string(text)); err != nil {
			fmt.Printf("*** Error reading configuration file: %v\n", err)
			os.Exit(1)
		} else {
			// Convert the file data to JSON and trim any leading/trailing whitespace
			text = jsonc.ToJSON(bytes.TrimSpace([]byte(data)))

			// test for JSON string, which must start with a '{' and not empty
			if len(text) == 0 || text[0] != '{' {
				fmt.Printf("*** string does not appear to be a valid JSON text missing starting '{'")
				os.Exit(1)
			}
		}
	}

	// Read the file and parse it
	if err := pCfg.parseConfig(text); err!= nil {
		fmt.Printf("Error parsing JSON data: %v\n", err)
		os.Exit(1)
	}
	return pCfg, nil
}

func (pc *PktgenConfig) parseConfig(text []byte) error {

	// Unmarshal json text into the Config structure
	if err := json.Unmarshal(text, pc); err != nil {
		return err
	}

	for core, mapping := range pc.Pktgen.Mappings {
		if mapping.Core >= gpc.MaxLogicalCores {
			return fmt.Errorf("invalid core ID: %d", mapping.Core)
		}
		mapping.Core = gpc.CoreID(core)
		if mapping.Mode.Value() == gpc.MainMode {
			mapping.Port = gpc.MaxEtherPorts
		}
	}

	return pc.validateConfig()
}

// readFile by passing in a filename or path to a JSON-C or JSON configuration
func (pc *PktgenConfig) readFile(s string) ([]byte, error) {
	if b, err := os.ReadFile(s); err != nil {
		return nil, err
	} else {
		return b, nil
	}
}

func (pc *PktgenConfig) CoreString() string {

	corelist := ""

	cores := make(sort.IntSlice, 0)
	for _, m := range pc.Pktgen.Mappings {
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

func (pc *PktgenConfig) NumChannels() uint16 {

	return pc.DPDK.NumChannels
}

func (pc *PktgenConfig) NumRanks() uint16 {

	return pc.DPDK.NumRanks
}

func (pc *PktgenConfig) MemorySize() uint64 {

	return pc.DPDK.MemorySize // Size of memory is in Bytes
}

func (pc *PktgenConfig) InMemory() bool {

	return pc.DPDK.InMemory
}

func (pc *PktgenConfig) FilePrefix() string {

	return pc.DPDK.FilePrefix
}

func (pc *PktgenConfig) PciList() []string {

	return pc.Pktgen.Cards
}

func (pc *PktgenConfig) CardCount() int {
	return len(pc.Pktgen.Cards)
}

func (pc *PktgenConfig) BaseMappings() []*BaseMapping {

	mappings := make([]Mapping, 0)
	for _, m := range pc.Pktgen.Mappings {
		mappings = append(mappings, m)
	}
	sort.Slice(mappings, func(i, j int) bool {
        return mappings[i].Core < mappings[j].Core
    })

	maps := make([]*BaseMapping, 0)
	for _, m := range mappings {
		m := &BaseMapping{
            Mode: m.Mode,
			Core: m.Core,
            Port: func() gpc.PortID {
				if m.Mode.Value() == gpc.MainMode {
					return gpc.MaxEtherPorts
				}
				return m.Port
			}(),
		}
		maps = append(maps, m)
	}

	return maps
}

func (pc *PktgenConfig) MappingCount() int {

	return len(pc.Pktgen.Mappings)
}

func (pc *PktgenConfig) LogTTY() int {

	return pc.PseudoTTY
}

func (pc *PktgenConfig) SetLogTTY(ptty int) {

	pc.PseudoTTY = ptty
}

func (pc *PktgenConfig) Promiscuous() bool {
	return pc.Pktgen.Promiscuous
}

// Return the command-line arguments for starting DPDK.
func (pc *PktgenConfig) GetArgsDPDK() ([]string, error) {

	args := []string{}
	if coreStr := pc.CoreString(); coreStr == "" {
		return nil, fmt.Errorf("core-list option is required")
	} else {
		args = append(args, "-l", coreStr)
	}
	if chnls := pc.NumChannels(); chnls > 0 {
		args = append(args, "-n", strconv.Itoa(int(chnls)))
	} else {
		return nil, fmt.Errorf("num-channels option is required")
	}
	if ranks := pc.NumRanks(); ranks > 0 {
		args = append(args, "-r", strconv.Itoa(int(ranks)))
	}
	if mem := pc.MemorySize(); mem > 0 {
		args = append(args, "-m", strconv.FormatUint(mem, 10))
	}
	if pc.InMemory() {
		args = append(args, "--in-memory")
	}
	if pc.FilePrefix() != "" {
		args = append(args, "--file-prefix", pc.FilePrefix())
	}

	if len(pc.PciList()) > 0 {
		for _, pci := range pc.PciList() {
			args = append(args, "-a", pci)
		}
	} else {
		return nil, fmt.Errorf("pci-list option is required")
	}

	return args, nil
}

func (pc *PktgenConfig) validateConfig() error {

	if pc.CardCount() == 0 {
		return fmt.Errorf("card list is empty")
	}
	if pc.MappingCount() == 0 {
		return fmt.Errorf("port mapping list is empty")
	}
	return nil
}

func (pc *PktgenConfig) String() string {

	if data, err := json.MarshalIndent(pc, "", "  "); err != nil {
		return fmt.Sprintf("error marshalling JSON: %v", err)
	} else {
		return string(data)
	}
}
