/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2022-2024 Intel Corporation.
 */

// Package pktgen is to create a traffic generator using DPDK as the I/O engine and Go
// frontend. Displaying the information in a clean and readable format.
package constants

import (
	"net"
)

type PacketConfig struct {
	PortIndex        int              // Port Index of the single packet
	TxCount          uint64           // Number of packets 0 == Forever
	PercentRate      float64          // Percent rate of packets per second
	PktSize          uint16           // Packet size
	BurstCount       uint16           // Size of packet burst
	TimeToLive       uint16           // Time to live value
	SrcPort, DstPort uint16           // Source and Destination port
	PType, ProtoType string           // Protocol type i.e., IPv4/TCP or UDP
	VlanId           uint16           // Vlan identifier
	SrcIP, DstIP     net.IPNet        // Source and Destination IP addresses
	SrcMAC, DstMAC   net.HardwareAddr // Source and Destination MAC addresses
	TxState          bool             // True is sending traffic
}
