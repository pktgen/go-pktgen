// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package gpcommon

const (
	DPDKProgramName               = "gpkt_dpdk" // Program name for DPDK
	MaxEtherPorts                 = 8           // Maximum number of Ethernet ports
	MaxLogicalCores               = 256         // Maximum number of logical cores
	InvalidQueueID                = 0xFFFF      // Invalid queue ID
	InvalidCacheSize              = 0xFFFF      // Invalid cache size
	DefaultNumChannels            = 0           // Default number of memory channels
	DefaultNumRanks               = 0           // Default number of memory ranks
	DefaultMemorySize             = 256         // Default memory size in MBytes
	DefaultInMemory               = false       // Default to use in-memory mode
	DefaultFilePrefix             = ""          // Default file prefix for file-based memory names
	DefaultCacheSize              = 256         // Default cache size
	DefaultRxDescSize             = 1024        // Default size of receive descriptor ring
	DefaultTxDescSize             = 2048        // Default size of transmit descriptor ring
	DefaultRxBurstSize            = 256         // Default receive burst size
	DefaultTxBurstSize            = 128         // Default transmit burst size
	DefaultMbufsPerPort           = (8 * 1024)  // Default number of mbufs per port
	DefaultPromiscuousMode        = true        // Default promiscuous mode
	EtherStatsQueueCntrs          = 16          // Number of Ethernet statistics queue counters
	CacheLineSize                 = 64          // Cache line size in bytes
	PortInfoNameSize              = 32          // Maximum size of port information name
	Million                uint64 = 1000000     // One million
	FrameOverheadSize      uint64 = 24          // Frame overhead size in bytes, includes FCS
	MinFrameSize           uint64 = 60          // Minimum frame size in bytes
	MaxFrameSize           uint64 = 1518        // Maximum frame size in bytes
	MaxJumboFrameSize      uint64 = 9000        // Maximum jumbo frame size in bytes
	OneGigaBits            uint64 = 1000000000

	DefaultLibraryPath = "./c-lib/usr/local/lib/x86_64-linux-gnu/"
	ChannelDPDKName    = "DPDK"
)

var ModeList []string = []string{"Unknown", "Main", "RxOnly", "TxOnly", "Rx/Tx"}

const (
	UnknownMode CoreMode = iota // Unknown core mode
	MainMode                    // Main core mode, used for processing control plane traffic
	RxMode                      // Receive only mode
	TxMode                      // Transmit only mode
	RxTxMode                    // Receive and transmit mode
)

const (
	UnknownMsgType = iota
	ExitMsgType
	LaunchMsgType
	PortMsgType
	MaxMsgTypes
)
