/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2022-2024 Intel Corporation.
 */

package fserde

import (
	"fmt"
)

// LayerHdr is the header for each layer containing common fields
type LayerHdr struct {
	fr        *Frame    // Pointer to the frame that contains the layer
	layerName LayerName // Name of the layer in display format.
	layerType LayerType // Type of the layer LayerEtherType, LayerDot1QType, LayerQinQType, ...
	proto     ProtoInfo // Protocol information for the layer.
}

// String returns a string representation of the layer header
func (ln *LayerHdr) String() string {
	return fmt.Sprintf("layerType=%+v, layerName=%v, proto=%+v, frame=%p",
		ln.layerType, ln.layerName, ln.proto, ln.fr)
}

// LayerConstructor is a function that creates a new layer header
func LayerConstructor(fr *Frame, layerName LayerName, layerType LayerType) *LayerHdr {
	return &LayerHdr{
		fr:        fr,
		layerName: layerName,
		layerType: layerType,
		proto:     ProtoInfo{},
	}
}

// Structure to hold all of the layer create functions
type NewFuncs struct {
	countNewFn    func(fr *Frame) *CountLayer
	dot1adNewFn   func(fr *Frame) *Dot1adLayer
	dot1qNewFn    func(fr *Frame) *Dot1qLayer
	echoNewFn     func(fr *Frame) *EchoLayer
	etherNewFn    func(fr *Frame) *EtherLayer
	icmpv4NewFn   func(fr *Frame) *ICMPv4Layer
	icmpv6NewFn   func(fr *Frame) *ICMPv6Layer
	ipv4NewFn     func(fr *Frame) *IPv4Layer
	ipv6NewFn     func(fr *Frame) *IPv6Layer
	payloadNewFn  func(fr *Frame) *PayloadLayer
	qinqNewFn     func(fr *Frame) *QinQLayer
	sctpNewFn     func(fr *Frame) *SCTPLayer
	tcpNewFn      func(fr *Frame) *TCPLayer
	tscNewFn      func(fr *Frame) *TSCLayer
	udpNewFn      func(fr *Frame) *UDPLayer
	vxlanNewFn    func(fr *Frame) *VxLanLayer
	defaultsNewFn func(fr *Frame) *DefaultsLayer
}

var newFuncs NewFuncs

func init() {

	// Register the layer create functions to the global structure
	newFuncs = NewFuncs{
		countNewFn:    CountNew,
		dot1adNewFn:   Dot1adNew,
		dot1qNewFn:    Dot1qNew,
		echoNewFn:     EchoNew,
		etherNewFn:    EtherNew,
		icmpv4NewFn:   ICMPv4New,
		icmpv6NewFn:   ICMPv6New,
		ipv4NewFn:     IPv4New,
		ipv6NewFn:     IPv6New,
		payloadNewFn:  PayloadNew,
		qinqNewFn:     QinQNew,
		sctpNewFn:     SCTPNew,
		tcpNewFn:      TCPNew,
		tscNewFn:      TSCNew,
		udpNewFn:      UDPNew,
		vxlanNewFn:    VxLanNew,
		defaultsNewFn: DefaultsNew,
	}
}
