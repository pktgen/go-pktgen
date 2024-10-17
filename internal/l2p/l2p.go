// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2019-2023 Intel Corporation

package l2p

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pktgen/go-pktgen/internal/tlog"
)

type CoreMode uint16 // Port mode (UnknownMode, RxMode, TxMode, RxTxMode)
const (
	UnknownMode CoreMode = iota // Unknown mode
	RxMode                      // Receive only mode
	TxMode                      // Transmit only mode
	RxTxMode                    // Receive and transmit mode
)

const (
	MAX_ETHER_PORTS   = 32  // Maximum number of Ethernet ports
	MAX_LOGICAL_CORES = 256 // Maximum number of logical cores
)

type Lcore uint16 // Logical core ID
type Lport uint16 // Logical port ID

type Port struct {
	Pid         Lport  // Port ID
	NumRxQueues uint16 // Number of receive queues
	NumTxQueues uint16 // Number of transmit queues
	PortInfo    any    // Port information pointer from C code
}

type LogicalPort struct {
	Mode     CoreMode // Mode (UnknownMode, RxMode, TxMode, RxTxMode)
	Lid      Lcore    // Logical core ID
	RxQid    uint16   // Number of receive queue ID
	TxQid    uint16   // Number of transmit queue ID
	PortInfo *Port    // Port information pointer
}

type L2p struct {
	MapList []string               // List of mapping entries
	LPorts  map[Lcore]*LogicalPort // Map of port IDs to LogicalPort structures
	Ports   map[Lport]*Port        // Map of port IDs to Port structures
}

func (l *L2p) AddMapping(pMap ...string) {
	for _, m := range pMap {
		l.MapList = append(l.MapList, m)
	}
}

func New() *L2p {
	return &L2p{
		MapList: make([]string, 0),
		LPorts:  make(map[Lcore]*LogicalPort),
		Ports:   make(map[Lport]*Port),
	}
}

func (l *L2p) ProcessMaps() error {

	for _, m := range l.MapList {
		// Process the mapping here.
		l.processMap(m)
	}

	return nil
}

/**
 * Parse the command line argument for port configuration
 *
 * DESCRIPTION
 * Parse the command line argument for port configuration.
 *
 * BNF: (or kind of BNF)
 *      <matrix-string> := """ <lcore-port> { "," <lcore-port>} """
 *		<lcore-port>	:= <lcore-list> "." <port>
 *		<lcore-list>	:= "[" <rx-list> ":" <tx-list> "]"
 *		<port>      	:= <num>
 *		<rx-list>		:= <num> { "/" (<num> | <list>) }
 *		<tx-list>		:= <num> { "/" (<num> | <list>) }
 *		<list>			:= <num>           { "/" (<range> | <list>) }
 *		<range>			:= <num> "-" <num> { "/" <range> }
 *		<num>			:= <digit>+
 *		<digit>			:= 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9
 *
 * BTW: A single lcore can only handle a single port/queue, which means
 *      you can not have a single core processing more then one network device or port.
 *
 *	1.0, 2.1, 3.2                 - core 1 handles port 0 rx/tx,
 *					                core 2 handles port 1 rx/tx
 *	[0-1].0, [2/4-5].1, ...		  - cores 0-1 handle port 0 rx/tx,
 *					                cores 2,4,5 handle port 1 rx/tx
 *	[1:2].0, [4:6].1, ...		  - core 1 handles port 0 rx,
 *					                core 2 handles port 0 tx,
 *	[1:2-3].0, [4:5-6].1, ...	  - core 1 handles port 1 rx, cores 2,3 handle port 0 tx
 *					                core 4 handles port 1 rx & core 5,6 handles port 1 tx
 *	[1-2:3].0, [4-5:6].1, ...	  - core 1,2 handles port 0 rx, core 3 handles port 0 tx
 *					                core 4,5 handles port 1 rx & core 6 handles port 1 tx
 *	[1-2:3-5].0, [4-5:6/8].1, ... - core 1,2 handles port 0 rx, core 3,4,5 handles port 0 tx
 *					                core 4,5 handles port 1 rx & core 6,8 handles port 1 tx
 *	BTW: you can use "{}" instead of "[]" or none at all as it does not matter to the syntax.
 *
 * RETURNS: N/A
 *
 * SEE ALSO:
 */
func (l *L2p) processMap(m string) error {

	tlog.DoPrintf("Processing L2P map %q\n", m)

	// Process the portmap into Cores and Port value
	fields := strings.Split(m, ".")
	if len(fields) != 2 {
		return fmt.Errorf("invalid L2P map format no corelist.port: %q", m)
	}
	fields[0] = strings.Trim(fields[0], "[]{}") // trim off the brackets if present

	v, err := strconv.ParseUint(fields[1], 10, 16) // Convert port number to integer
	if err != nil {
		return fmt.Errorf("invalid port number: %q", fields[1])
	}
	pid := Lport(v)
	if pid > MAX_ETHER_PORTS {
		return fmt.Errorf("invalid port number: %v", pid)
	}
	// Check if port already exists
	if _, ok := l.Ports[pid]; ok {
		return fmt.Errorf("port %v already used", pid)
	}
	port := &Port{Pid: pid}
	l.Ports[pid] = port

	cores := strings.Split(fields[0], ":") // Split into Rx and Tx cores

	switch len(cores) {
	case 2:
		if err := l.processCores(port, cores[0], RxMode); err != nil {
			return err
		}
		if err := l.processCores(port, cores[1], TxMode); err != nil {
			return err
		}
	case 1:
		if err := l.processCores(port, cores[0], RxTxMode); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid L2P map format: %q", m)
	}

	return nil
}

func (l *L2p) processCores(port *Port, coreStr string, mode CoreMode) error {

	var low, high Lcore
	var err error

	cores := strings.Split(coreStr, ",")
	for _, core := range cores {
		if low, high, err = parseCoreRange(core); err != nil {
			return err
		} else {
			for i := low; i <= high; i++ {
				lport := l.addLogicalPort(i, port, mode)
				switch mode {
				case RxMode:
					lport.RxQid = port.NumRxQueues
					port.NumRxQueues++

				case TxMode:
					lport.TxQid = port.NumTxQueues
					port.NumTxQueues++
				case RxTxMode:
					lport.RxQid = port.NumRxQueues
					lport.TxQid = port.NumTxQueues
					port.NumRxQueues++
					port.NumTxQueues++
				default:
					return fmt.Errorf("invalid RxTx mode: %v", mode)
				}
			}
		}
	}

	return nil
}

func parseCoreRange(s string) (Lcore, Lcore, error) {
	var low, high Lcore

	parts := strings.Split(s, "-") // Split on '-' if range is present

	switch len(parts) {
	case 2: // Range of cores
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid range start: %q", parts[0])
		}
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid range end: %q", parts[1])
		}
		low, high = Lcore(start), Lcore(end)
	case 1: // Single core
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid range start: %q", parts[0])
		}
		low, high = Lcore(start), Lcore(start)
	default: // Invalid range format
		return 0, 0, fmt.Errorf("invalid range format: %q", s)
	}

	if low > high {
		return 0, 0, fmt.Errorf("invalid core range: %s", s)
	} else if high > MAX_LOGICAL_CORES {
		return 0, 0, fmt.Errorf("invalid core number: %d", high)
	}

	return low, high, nil
}

func (l *L2p) addLogicalPort(core Lcore, port *Port, mode CoreMode) *LogicalPort {
	if lport, ok := l.LPorts[core]; !ok {
		lport := &LogicalPort{
			Mode:     mode,
			Lid:      core,
			PortInfo: port,
		}
		l.LPorts[core] = lport
		return lport
	} else {
		return lport
	}
}

func (l *L2p) PortList() []Port {

	list := make([]Port, 0)
	for _, p := range l.Ports {
		list = append(list, *p)
	}
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Pid < list[j].Pid
	})

	return list
}

func (l *L2p) LogicalPortList() []LogicalPort {

	list := make([]LogicalPort, 0)
	for _, p := range l.LPorts {
		list = append(list, *p)
	}
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].PortInfo.Pid < list[j].PortInfo.Pid
	})

	return list
}

func (l *L2p) SetPortInfo(pid Lport, info any) error {

	if port, ok := l.Ports[pid]; ok {
		port.PortInfo = info
		return nil
	}
	return fmt.Errorf("port %v not found", pid)
}

// (CoreMode)MarshalJSON decodes JSON value into string.
func (cm CoreMode) MarshalJSON() ([]byte, error) {
	switch cm {
	default:
		return []byte(`"UnknownMode"`), nil
	case RxMode:
		return []byte(`"RxOnlyMode"`), nil
	case TxMode:
		return []byte(`"TxOnlyMode"`), nil
	case RxTxMode:
		return []byte(`"RxTxMode"`), nil
	}
}

func (l *L2p) Marshal() string {

	if b, err := json.MarshalIndent(l, "", "  "); err != nil {
		return ""
	} else {
		return string(b)
	}
}
