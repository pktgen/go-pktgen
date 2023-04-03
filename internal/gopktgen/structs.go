// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package gopktgen

import (
	"unsafe"

	gpc "github.com/pktgen/go-pktgen/internal/gpcommon"
)

type (
	PhysicalPort struct {
		Pid         gpc.PortID // Port ID
		NumRxQueues uint16     // Number of receive queues
		NumTxQueues uint16     // Number of transmit queues
	}

	LogicalPort struct {
		Port          *PhysicalPort // Physical port structure pointer
		LogicalPortID gpc.LPortID   // Logical port ID i.e., ((PortID << 16) | QueueID)
		RxQid         gpc.QueueID   // Receive queue ID
		TxQid         gpc.QueueID   // Transmit queue ID}
	}

	LogicalCore struct {
		LPort *LogicalPort // Logical port structure pointer
		Mode  gpc.CoreMode // Mode (UnknownMode, RxMode, TxMode, RxTxMode or DisplayMode)
		Core  gpc.CoreID   // Core ID or logical core ID
	}

	L2p struct {
		Cores       map[gpc.CoreID]*LogicalCore  // Map of CoreIDs to CoreInfo structures
		LPorts      map[gpc.LPortID]*LogicalPort // Map of LogicalPortIDs to LogicalPort structures
		Ports       map[gpc.PortID]*PhysicalPort // Map of PortIDs to PhysicalPort structures
		BaseMapList []*BaseMapping               // List of mapping entries
		PortCount   uint16                       // Number of ports
	}

	// L2pConfig represents the configuration for the L2p structure.
	L2pConfig struct {
		LPortID     gpc.LPortID  // LPort ID
		CoreID      gpc.CoreID   // Core ID
		CoreMode    gpc.CoreMode // Core mode
		RxQid       gpc.QueueID  // RxQid
		TxQid       gpc.QueueID  // TxQid
		PortID      gpc.PortID   // Port ID
		NumRxQueues uint16       // Number of Rx queues
		NumTxQueues uint16       // Number of Tx queues
	}

	// Go-Pktgen is a software packet generator for Intel Corp NICs.
	// This configuration file (pktgen.jsonc) contains various settings for starting
	// and configuring DPDK started by Go-Pktgen. Not all of the DPDK options are supported
	// in the Go-Pktgen configuration file.
	// The JSONC format allows comments and supports basic JSON syntax.
	// The JSONC format is converted to JSON format before parsing the structured information.
	// Most of the JSONC data is simple to parse and understood, except for some complex data types.
	// Global Options:
	// 	"log-tty"         - field supports a single integer value or tty number from /dev/pts/X, e.g., "3".
	// DPDK Options:
	// 	"num-channels"    - number of memory channels, field supports a positive integer value, e.g., default 0.
	// 	"num-ranks"       - number of memory ranks, field supports a positive integer value, e.g., default 0.
	// 	"memory-size"     - field supports a positive integer value in MBytes, e.g., default 256.
	// 	"in-memory"       - field supports a boolean value to have DPDK not use file base memory, e.g.,  default "false".
	// 	"file-prefix"     - field supports a string prefix value for the file based memory names, e.g., "pktgen_" or "rte_map_".
	// 	"rx-desc-size"    - Optional, Size of RX descriptor ring, zero use default
	// 	"tx-desc-size"    - Optional, Size of TX descriptor ring, zero use default
	// 	"rx-burst_size"   - Optional, Size of Rx burst size, zero use default
	// 	"Tx-burst_size"   - Optional, Size of Tx burst size, zero use default
	// Pktgen Options:
	// 	The "cards"       - List of NIC card PCIe addresses.
	// 	The "mapping"     - List of core and port mappings. Each entry should be in the format:
	// 	The "promiscuous" - field supports a boolean value, e.g., default "true".
	//
	PktgenConfig struct {
		PseudoTTY int           `json:"log-tty"`
		DPDK      OptionsDPDK   `json:"dpdk"`
		Pktgen    OptionsPktgen `json:"pktgen"`
	}

	OptionsDPDK struct {
		FilePrefix  string `json:"file-prefix"`
		MemorySize  uint64 `json:"memory-size"` // in MBytes
		NumChannels uint16 `json:"num-channels"`
		NumRanks    uint16 `json:"num-ranks"`
		InMemory    bool   `json:"in-memory"`
	}

	BaseMapping struct {
		Mode gpc.ModeString `json:"mode"`
		Core gpc.CoreID     `json:"core"`
		Port gpc.PortID     `json:"port"`
	}

	Mapping struct {
		BaseMapping
		RxDesc     int16 `json:"rx-desc"`
		TxDesc     int16 `json:"tx-desc"`
		RxBurst    int16 `json:"rx-burst"`
		TxBurst    int16 `json:"tx-burst"`
		CacheSize  int16 `json:"cache-size"`
		NumPackets int32 `json:"num-packets"`
	}

	OptionsPktgen struct {
		Cards       []string  `json:"cards"`
		Mappings    []Mapping `json:"mappings"`
		Promiscuous bool      `json:"promiscuous"`
	}

	GoPktgenOption func(*GoPktgen)

	LibInfo struct {
		Name   string  // Name of the library.
		Handle uintptr // library handle from dlopen().
	}

	ApiInfo struct {
		FuncName string // Function Name of the library.
		FuncPtr  any    // Function pointer.
	}

	// GoPktgen API functions
	GoPktgenApi struct {
		AddArgv         func(arg string) int                               // Add a argv value
		L2pConfig       func(cfg unsafe.Pointer) int                       // Add L2p configuration
		Startup         func(log_path string) int                          // Start function with log_path
		Shutdown        func() int                                         // Stop function
		PortSetInfo     func(portCfg *gpc.PortConfig) int                  // Set port information
		PortGetInfo     func(portID gpc.PortID) *gpc.PortConfig            // Get port information
		PortFreeInfo    func(cfg *gpc.PortConfig)                          // Free port information structure
		PortEtherStats  func(portID gpc.PortID, stats unsafe.Pointer) int  // Get port statistics
		PortPacketStats func(portID gpc.PortID, stats unsafe.Pointer) int  // Get port statistics
		PortLinkStatus  func(portID gpc.PortID) uint64                     // Get link status encoded as a uint64
		PortMacAddress  func(portID gpc.PortID, mac unsafe.Pointer) int    // Get port MAC address
		PortDeviceInfo  func(portID gpc.PortID, info unsafe.Pointer) int   // Get port device information
		ChannelCreate   func(name string, size uint32) uintptr             // MsgChan initialize
		ChannelAttach   func(name string) uintptr                          // MsgChan attach
		ChannelDestroy  func(mc uintptr) int                               // MsgChan destroy
		ChannelRecv     func(mc uintptr, data unsafe.Pointer, len int) int // MsgChan receive data burst
		ChannelSend     func(mc uintptr, data unsafe.Pointer, len int) int // MsgChan send data burst
	}

	GoPktgen struct {
		GoPktgenApi
		pCfg      *PktgenConfig   // Configuration system instance
		l2p       *L2p            // L2p instance
		libList   []*LibInfo      // Slice of libraries names to their handles
		portData  []*gpc.PortData // List of port information structures
		portStats []*PortStats    // List of port statistics structures
		basePath  string          // Base path for libraries
		logPath   string          // Log path for tlog_printf function
		dpdkChan  uintptr         // Message channel for GoPktgen to receive and process messages
	}
)
