/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2023-2025 Intel Corporation.
 */

package fserde

import (
	"strings"
)

const (
	EtherTypeIPv4   = 0x0800 // IPv4 EtherType value
	EtherTypeIPv6   = 0x86dd // IPv6 EtherType value
	EtherTypeARP    = 0x0806 // ARP EtherType value
	Dot1QID         = 0x8100 // IEEE 802.1Q VLAN ID EtherType value
	QinQID          = 0x88a8 // IEEE 802.1Q QinQ VLAN ID EtherType value
	HardwareAddrLen = 6      // Length of hardware MAC address
	MinPacketLen    = 60     // Minimum Ethernet frame length without FCS
	MaxPacketLen    = 1514   // Maximum Ethernet frame length without FCS
	ProtocolUDP     = 17     // UDP protocol number
	ProtocolTCP     = 6      // TCP protocol number
	ProtocolIPv4    = 4      // IPv4 protocol number
	ProtocolIPv6    = 41     // IPv6 protocol number
	ProtocolICMPv4  = 1      // ICMPv4 protocol number
	ProtocolICMPv6  = 58     // ICMPv6 protocol number
)

type LayerType int    // Layer type index value
type LayerName string // Layer type name

const (
	// Index values for each layer, must match the list below set of constants.
	LayerEtherType LayerType = iota
	LayerDot1QType
	LayerQinQType
	LayerDot1ADType
	LayerIPv4Type
	LayerIPv6Type
	LayerTCPType
	LayerUDPType
	LayerICMPv4Type
	LayerICMPv6Type
	LayerSCTPType
	LayerVxLanType
	LayerEchoType
	LayerTSCType
	LayerPayloadType
	LayerDefaultsType
	LayerCountType
	MaxLayerType
)

// String returns the string representation of the layer name string
func (l LayerType) String() LayerName {
	return LayerNames[l]
}

const (
	// These strings are used for displaying layer names.
	// Must match the order of the layers above.
	LayerEther    LayerName = "Ether"
	LayerDot1Q    LayerName = "Dot1Q"
	LayerQinQ     LayerName = "QinQ"
	LayerDot1AD   LayerName = "Dot1AD"
	LayerIPv4     LayerName = "IPv4"
	LayerIPv6     LayerName = "IPv6"
	LayerTCP      LayerName = "TCP"
	LayerUDP      LayerName = "UDP"
	LayerICMPv4   LayerName = "ICMPv4"
	LayerICMPv6   LayerName = "ICMPv6"
	LayerSCTP     LayerName = "SCTP"
	LayerVxLan    LayerName = "VxLan"
	LayerEcho     LayerName = "Echo"
	LayerTSC      LayerName = "TSC"
	LayerPayload  LayerName = "Payload"
	LayerDefaults LayerName = "Defaults"
	LayerCount    LayerName = "Count"
	LayerDone     LayerName = "Done"
)

var LayerNames = [...]LayerName{
	LayerEther,
	LayerDot1Q,
	LayerQinQ,
	LayerDot1AD,
	LayerIPv4,
	LayerIPv6,
	LayerTCP,
	LayerUDP,
	LayerICMPv4,
	LayerICMPv6,
	LayerSCTP,
	LayerVxLan,
	LayerEcho,
	LayerTSC,
	LayerPayload,
	LayerDefaults,
	LayerCount,
	LayerDone,
}

// layerTypeFromNames returns the layer type from the layer name string
func layerTypeFromName(name string) LayerType {

	for i := range LayerNames {
		if strings.EqualFold(name, string(LayerNames[i])) {
			return LayerType(i)
		}
	}
	return MaxLayerType
}

// findLayerName returns the layer type from the layer name string
func findLayerName(lName string) LayerName {
	for i := range LayerNames {
		if strings.EqualFold(lName, string(LayerNames[i])) {
			return LayerNames[i]
		}
	}
	return LayerName("LayerName-Unknown")
}

const (
	ProtoL2 = iota
	ProtoL3
	ProtoL4
	ProtoL5
	ProtoL6
	ProtoL7
	MaxProtocols
)
