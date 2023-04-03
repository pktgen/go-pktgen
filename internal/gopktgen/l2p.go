// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package gopktgen

import (
	"encoding/json"
	"fmt"
	"sort"

	gpc "github.com/pktgen/go-pktgen/internal/gpcommon"
)

func (g *GoPktgen) AddMapping(pMap ...*BaseMapping) {

	l := g.l2p
	for _, m := range pMap {
		fmt.Printf("Processing mapping: %+v\n", m)

		l.BaseMapList = append(l.BaseMapList, m)
	}
}

func (g *GoPktgen) ProcessMaps() error {

	l := g.l2p
	for _, m := range l.BaseMapList {
		// Process the mapping here.
		if err := g.processMap(m); err != nil {
			return err
		}
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
func (g *GoPktgen) processMap(m *BaseMapping) error {

	var lcore *LogicalCore
	var port *PhysicalPort

	l := g.l2p

	if m.Core >= gpc.MaxLogicalCores {
		return fmt.Errorf("invalid core number: %v", m.Core)
	}
	if m.Mode.Value() < gpc.MainMode || m.Mode.Value() > gpc.RxTxMode {
		return fmt.Errorf("Invalid mode %v", m.Mode)
	}
	if m.Port > gpc.MaxEtherPorts {
		return fmt.Errorf("invalid port number: %v", m.Port)
	}

	// Check if core already exists
	lcore, ok := l.Cores[m.Core]
	if ok {
		return fmt.Errorf("core %v already used", lcore.Core)
	} else {
		lcore = &LogicalCore{Core: m.Core, Mode: m.Mode.Value()}
		l.Cores[m.Core] = lcore
	}

	// Check if port already exists and mode is valid
	if lcore.Mode != gpc.MainMode && lcore.Mode != gpc.UnknownMode {
		port, ok = l.Ports[m.Port]
		if ok {
			return fmt.Errorf("port %v mode {%v} already used", port.Pid, lcore.Mode)
		} else {
			port = &PhysicalPort{Pid: gpc.PortID(m.Port)}

			l.Ports[m.Port] = port
			l.PortCount++
		}
	} else {
		port = &PhysicalPort{Pid: gpc.MaxEtherPorts}
	}

	// Add ports to the L2p structure map
	lport := g.newLogicalPort(port, m.Mode.Value())
	lcore.LPort = lport

	switch m.Mode.Value() {
	case gpc.MainMode:
		// skip main core
	case gpc.RxMode:
		lport.RxQid = gpc.QueueID(port.NumRxQueues)
		port.NumRxQueues++

	case gpc.TxMode:
		lport.TxQid = gpc.QueueID(port.NumTxQueues)
		port.NumTxQueues++

	case gpc.RxTxMode:
		lport.RxQid = gpc.QueueID(port.NumRxQueues)
		lport.TxQid = gpc.QueueID(port.NumTxQueues)
		port.NumRxQueues++
		port.NumTxQueues++

	default:
		return fmt.Errorf("invalid RxTx mode: %v", m.Mode)
	}

	return nil
}

func (g *GoPktgen) newLogicalPort(port *PhysicalPort, mode gpc.CoreMode) *LogicalPort {
	var qid gpc.QueueID

	// use mode to determine queue ID to use for logical port ID
	switch mode {
	case gpc.RxMode, gpc.RxTxMode:
		qid = gpc.QueueID(port.NumRxQueues)
	case gpc.TxMode:
		qid = gpc.QueueID(port.NumTxQueues)
	}
	// Create a new Logical Port ID from the Physical Port ID and queue ID
	logicalPort := gpc.ToLogicalPort(port.Pid, qid)

	// Check if logical port already exists
	if lport, ok := g.l2p.LPorts[logicalPort]; !ok {
		lport := &LogicalPort{
			Port:          port,
			LogicalPortID: logicalPort,
			RxQid:         gpc.InvalidQueueID,
			TxQid:         gpc.InvalidQueueID,
		}
		g.l2p.LPorts[logicalPort] = lport
		return lport
	} else {
		return lport
	}
}

func (g *GoPktgen) PortList() []PhysicalPort {

	list := make([]PhysicalPort, 0)
	for _, p := range g.l2p.Ports {
		list = append(list, *p)
	}
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Pid < list[j].Pid
	})

	return list
}

func (g *GoPktgen) LogicalPortEntry(pid gpc.PortID) *LogicalPort {

	for _, p := range g.l2p.LPorts {
		if p.Port.Pid == pid {
			return p
		}
	}
	return nil
}

func (g *GoPktgen) LogicalPortList() []LogicalPort {

	list := make([]LogicalPort, 0)
	for _, p := range g.l2p.LPorts {
		list = append(list, *p)
	}
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Port.Pid < list[j].Port.Pid
	})

	return list
}

func (g *GoPktgen) CoreList() []LogicalCore {

	list := make([]LogicalCore, 0)
	for _, c := range g.l2p.Cores {
		list = append(list, *c)
	}
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Core < list[j].Core
	})

	return list
}

func (g *GoPktgen) Core(coreID gpc.CoreID) *LogicalCore {
	if ci, ok := g.l2p.Cores[coreID]; ok {
		return ci
	}
	return nil
}

func (g *GoPktgen) PortCount() uint16 {
	return g.l2p.PortCount
}

// (L2p) NumQueues returns the number of Rx and Tx queues for a given port.
func (g *GoPktgen) NumQueues(pid gpc.PortID) (uint16, uint16) {

	if port, ok := g.l2p.Ports[pid]; ok {
		return port.NumRxQueues, port.NumTxQueues
	} else {
		return 0, 0
	}
}

func (g *GoPktgen) Marshal() string {

	if b, err := json.MarshalIndent(g.l2p.Cores, "", "  "); err != nil {
		fmt.Printf("*** %v\n", err)
		return ""
	} else {
		return string(b)
	}
}
