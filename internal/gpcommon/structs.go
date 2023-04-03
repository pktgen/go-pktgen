// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package gpcommon

import (
	"encoding/json"
	"fmt"
	"net"
)

type (
	ModeString string // Core mode (Unknown, Rx, Tx, RxTxMode)
	CoreMode   uint16 // Port mode (UnknownMode, RxMode, TxMode, RxTxMode)
	CoreID     uint16 // Core ID
	PortID     uint16 // Port ID
	QueueID    uint16 // Queue ID
	LPortID    uint32 // Logical = Port/Queue ID ((PortID << 16) | QueueID)
	PciAddr    string // PCIe address of the NIC card.

	// ChannelMsg represents the data for a channel. Must match channel_msg_t in channel.h
	ChannelMsg struct {
		Action   uint16                  // Action for the channel
		Len      uint16                  // Length of the data
		reserved uint32                  // Reserved for future use
		Data     [CacheLineSize - 8]byte // Data to be sent or received
	}

	// IPAddress represents an IPv4 address.
	LinkState struct {
		Speed   uint32 // Link speed in Mbps
		Duplex  bool   // Link duplex (true = full-duplex, false = half-duplex)
		AutoNeg bool   // Auto-negotiation is enabled (true = enabled, false = disabled)
		Status  bool   // Link is up (true = up, false = down)
	}

	// MacAddress represents a MAC address.
	MacAddress struct {
		Address [6]byte // MAC address
	}

	// PortData represents the data for a port
	PortData struct {
		MacAddress MacAddress // MAC address
		PortID     PortID     // Port ID
		PCIeID     string     // PCIe ID
	}

	// EtherStats represents the statistics for a port must match DPDK.rte_eth_stats in ethdev.h
	EtherStats struct {
		IPackets uint64 // Total number of successfully received packets.
		OPackets uint64 // Total number of successfully transmitted packets.
		IBytes   uint64 // Total number of successfully received bytes.
		OBytes   uint64 // Total number of successfully transmitted bytes.
		IMissed  uint64 // Total number of Rx packets dropped by the HW.
		IErrors  uint64 // Total number of erroneous received packets.
		OErrors  uint64 // Total number of failed transmitted packets.
		RxNombuf uint64 // Total number of Rx mbuf allocation failures.

		// Per Queue statistics
		QIPackets [EtherStatsQueueCntrs]uint64 // Total number of queue Rx packets.
		QOPackets [EtherStatsQueueCntrs]uint64 // Total number of queue Tx packets.
		QIBytes   [EtherStatsQueueCntrs]uint64 // Total number of queue Rx queue bytes.
		QOBytes   [EtherStatsQueueCntrs]uint64 // Total number of queue Tx queue bytes.
		QErrors   [EtherStatsQueueCntrs]uint64 // Total number of queue dropped packets.
	}

	// PacketStats represents the statistics for packet types.
	PacketStats struct {
		Broadcast      uint64 // Number of broadcast packets received
		Multicast      uint64 // Number of multicast packets received
		Size64         uint64 // Number of 64-bytes packets received
		Size65To127    uint64 // Number of 65-127-byte packets received
		Size128To255   uint64 // Number of 128-255-byte packets received
		Size256To511   uint64 // Number of 256-511-byte packets received
		Size512To1023  uint64 // Number of 512-1023-byte packets received
		Size1024To1518 uint64 // Number of 1024-1518-byte packets received
		RuntPackets    uint64 // Number of Runt packets received
		JumboPackets   uint64 // Number of jumbo packets received
		ArpPackets     uint64 // Number of ARP packets received
		IcmpPackets    uint64 // Number of ICMP packets received
	}

	PortStatistics struct {
		Ether  *EtherStats  // Ethernet Hardware statistics
		Packet *PacketStats // Packet type statistics
	}

	// PortInfo represents information about a port.
	PortConfig struct {
		PortID       PortID // Port ID
		NumRxQueues  uint16 // Number of receive queues
		NumTxQueues  uint16 // Number of transmit queues
		RxDescSize   uint16 // Receive descriptor size
		TxDescSize   uint16 // Transmit descriptor size
		RxBurstSize  uint16 // Receive burst size
		TxBurstSize  uint16 // Transmit burst size
		CacheSize    uint16 // Cache size
		MbufsPerPort uint32 // Number of mbufs per port
	}

	PortDeviceInfo struct {
		Name            [PortInfoNameSize]byte // Device name
		BusName         [PortInfoNameSize]byte // Bus name
		MacAddr         MacAddress             // MAC address
		IfIndex         uint32                 // Interface index
		MinMtu          uint32                 // Minimum MTU
		MaxMtu          uint32                 // Maximum MTU
		MinRxBufSize    uint32                 // Minimum receive buffer size
		MaxRxBufSize    uint32                 // Maximum receive buffer size
		MaxRzPktLen     uint32                 // Maximum receive packet length
		MaxRxQueues     uint32                 // Maximum receive queues
		MaxTxQueues     uint32                 // Maximum transmit queues
		MaxMacAddrs     uint32                 // Maximum MAC addresses
		MaxHashMacAddrs uint32                 // Maximum hash MAC addresses
		MacVfs          uint32                 // Maximum MAC VFs
		NbRxQueues      uint32                 // Number of receive queues
		NbTxQueues      uint32                 // Number of transmit queues
		SocketID        uint32                 // Socket ID
	}

	PhysicalPort struct {
		Pid         PortID // Port ID
		NumRxQueues uint16 // Number of receive queues
		NumTxQueues uint16 // Number of transmit queues
	}

	LogicalPort struct {
		Port          *PhysicalPort // Physical port structure pointer
		LogicalPortID LPortID       // Logical port ID i.e., ((PortID << 16) | QueueID)
		RxQid         QueueID       // Receive queue ID
		TxQid         QueueID       // Transmit queue ID}
	}

	LogicalCore struct {
		LPort *LogicalPort // Logical port structure pointer
		Mode  CoreMode     // Mode (UnknownMode, RxMode, TxMode, RxTxMode or DisplayMode)
		Core  CoreID       // Core ID or logical core ID
	}

	L2p struct {
		Cores     map[CoreID]*LogicalCore  // Map of CoreIDs to CoreInfo structures
		LPorts    map[LPortID]*LogicalPort // Map of LogicalPortIDs to LogicalPort structures
		Ports     map[PortID]*PhysicalPort // Map of PortIDs to PhysicalPort structures
		MapList   []*Mapping               // List of mapping entries
		PortCount uint16                   // Number of ports
	}

	// L2pConfig represents the configuration for the L2p structure.
	L2pConfig struct {
		CoreID      CoreID   // Core ID
		CoreMode    CoreMode // Core mode
		LPortID     LPortID  // LPort ID
		RxQid       QueueID  // RxQid
		TxQid       QueueID  // TxQid
		PortID      PortID   // Port ID
		NumRxQueues uint16   // Number of Rx queues
		NumTxQueues uint16   // Number of Tx queues
	}

	/*
		Go-Pktgen is a software packet generator for Intel Corp NICs.

		This configuration file (pktgen.jsonc) contains various settings for starting
		and configuring DPDK started by Go-Pktgen. Not all of the DPDK options are supported
		in the Go-Pktgen configuration file.

		The JSONC format allows comments and supports basic JSON syntax.
		The JSONC format is converted to JSON format before parsing the structured information.

		Most of the JSONC data is simple to parse and understood, except for some complex data types.
		Global Options:

			"debug-tty"    - field supports a single integer value or tty number from /dev/pts/X, e.g., "3".

		DPDK Options:

			"num-channels"   - number of memory channels, field supports a positive integer value, e.g., default 0.
			"num-ranks"      - number of memory ranks, field supports a positive integer value, e.g., default 0.
			"memory-size"    - field supports a positive integer value in MBytes, e.g., default 256.
			"in-memory"      - field supports a boolean value to have DPDK not use file base memory, e.g.,  default "false".
			"file-prefix"    - field supports a string prefix value for the file based memory names, e.g., "pktgen_" or "rte_map_".
			"rx-desc-size"   - Optional, Size of RX descriptor ring, zero use default
			"tx-desc-size"   - Optional, Size of TX descriptor ring, zero use default
			"rx-burst_size"  - Optional, Size of Rx burst size, zero use default
			"Tx-burst_size"  - Optional, Size of Tx burst size, zero use default

		Pktgen Options:

			The "cards"        - List of NIC card PCIe addresses.
			The "mapping"      - List of core and port mappings. Each entry should be in the format:
			The "promiscuous"  - field supports a boolean value, e.g., default "true".
	*/
	PktgenConfig struct {
		LogTTY int           `json:"log-tty"`
		DPDK   OptionsDPDK   `json:"dpdk"`
		Pktgen OptionsPktgen `json:"pktgen"`
	}

	OptionsDPDK struct {
		NumChannels uint16 `json:"num-channels"`
		NumRanks    uint16 `json:"num-ranks"`
		MemorySize  uint64 `json:"memory-size"` // in MBytes
		InMemory    bool   `json:"in-memory"`
		FilePrefix  string `json:"file-prefix"`
	}

	Mapping struct {
		Mode       ModeString `json:"mode"`
		Core       CoreID     `json:"core"`
		Port       PortID     `json:"port"`
		RxDesc     int16      `json:"rx-desc"`
		TxDesc     int16      `json:"tx-desc"`
		RxBurst    int16      `json:"rx-burst"`
		TxBurst    int16      `json:"tx-burst"`
		CacheSize  int16      `json:"cache-size"`
		NumPackets int32      `json:"num-packets"`
	}

	OptionsPktgen struct {
		Cards       []string  `json:"cards"`
		Mappings    []Mapping `json:"mappings"`
		Promiscuous bool      `json:"promiscuous"`
	}
)

func NewPortConfig(portID PortID, numRx, numTx uint16) *PortConfig {
	return &PortConfig{
		PortID:       portID,
		NumRxQueues:  numRx,
		NumTxQueues:  numTx,
		RxDescSize:   DefaultRxDescSize,
		TxDescSize:   DefaultTxDescSize,
		RxBurstSize:  DefaultRxBurstSize,
		TxBurstSize:  DefaultTxBurstSize,
		CacheSize:    DefaultCacheSize,
		MbufsPerPort: DefaultMbufsPerPort,
	}
}

// String method for LinkStatus in "FD-40000-UP" Duplex-Speed-Status
func (l LinkState) String() string {

	str := ""
	if l.Duplex {
		str += "FD-"
	} else {
		str += "HD-"
	}
	str += fmt.Sprintf("%v-", l.Speed)
	if l.Status {
		str += "UP"
	} else {
		str += "DOWN"
	}
	return str
}

func (m MacAddress) String() string {
	return net.HardwareAddr(m.Address[:]).String()
}

func (lp LPortID) String() string {
	pid, qid := lp.FromLogicalPort()
	return fmt.Sprintf("LPortID(PID:%d, QID:%d)", pid, qid)
}

func (lp LPortID) FromLogicalPort() (CoreID, QueueID) {
	return CoreID((lp >> 16) & 0xFFFF), QueueID(lp & 0xFFFF)
}

func ToLogicalPort(pid PortID, qid QueueID) LPortID {
	return LPortID((uint32(pid) << 16) | uint32(qid))
}

func (cm CoreMode) String() string {

	return ModeList[cm]
}

func (ms ModeString) String() string {
	return ms.String()
}

func (cd *PktgenConfig) String() string {

	if data, err := json.MarshalIndent(cd, "", "  "); err != nil {
		return fmt.Sprintf("error marshalling JSON: %v", err)
	} else {
		return string(data)
	}
}

func (cd *PktgenConfig) ValidateConfig() error {

	if len(cd.Pktgen.Cards) == 0 {
		return fmt.Errorf("card list is empty")
	}
	if len(cd.Pktgen.Mappings) == 0 {
		return fmt.Errorf("port mapping list is empty")
	}
	return nil
}
