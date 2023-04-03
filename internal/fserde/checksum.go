/* SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2022-2024 Intel Corporation.
 */

package fserde

// Some code is from https://github.com/aregm/nff-go/blob/master/packet/checksum.go
// and modified to work within fserde.
//
// SPDX-License-Identifier: BSD-3-Clause
// Copyright 2017 Intel Corporation.

import (
	"encoding/binary"
	"net"

	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// SwapBytesUint16 swaps uint16 in Little Endian and Big Endian
func SwapUint16(x uint16) uint16 {
	return x<<8 | x>>8
}

// SwapBytesUint32 swaps uint32 in Little Endian and Big Endian
func SwapUint32(x uint32) uint32 {
	return ((x & 0x000000ff) << 24) | ((x & 0x0000ff00) << 8) |
		((x & 0x00ff0000) >> 8) | ((x & 0xff000000) >> 24)
}

func SwapIPv4Addr(x net.IP) net.IP {
	b := x.To4()
	return net.IPv4(b[0], b[1], b[2], b[3])
}

func reduceChecksum(sum uint32) uint16 {
	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	return uint16(sum)
}

// Calculates checksum of memory for a byte array and length.
func dataChecksum(data []byte, length int) uint32 {
	var sum uint32

	dbug.Printf("length: %d, data %v\n", length, data)
	if length == 0 {
		return 0
	}

	// Sum all 16-bit words in the header
	for i := 0; i < length-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}

	// If the header has an odd number of bytes, add the last byte
	if len(data)%2 != 0 {
		sum += uint32(data[len(data)-1]) << 8
	}

	return sum
}

// IPv4HeaderChecksum calculates checksum of IP header
func IPv4HeaderChecksum(hdr *ipv4.Header) uint16 {

	hl := hdr.Version<<4 | (hdr.Len >> 2)

	src := binary.BigEndian.Uint32(hdr.Src.To4())
	dst := binary.BigEndian.Uint32(hdr.Dst.To4())

	arr := make([]uint16, IPv4MinLen>>1) // 16-bit words for IP header length

	arr[0] = uint16(hl)<<8 | uint16(hdr.TOS)
	arr[1] = uint16(hdr.TotalLen)
	arr[2] = uint16(hdr.ID)
	arr[3] = uint16(hdr.FragOff)
	arr[4] = uint16(uint16(hdr.TTL<<8) | uint16(hdr.Protocol))
	arr[5] = uint16(0) // Checksum
	arr[6] = uint16(src >> 16)
	arr[7] = uint16(src & 0xFFFF)
	arr[8] = uint16(dst >> 16)
	arr[9] = uint16(dst & 0xFFFF)

	sum := uint32(0)
	for i := 0; i < len(arr); i++ {
		sum += uint32(arr[i])
	}
	sum += dataChecksum(hdr.Options, len(hdr.Options))

	return ^reduceChecksum(sum)
}

func PseudoHdrIPv4Checksum(ipSrc, ipDst net.IP, proto int, length uint16) uint32 {

	src := binary.BigEndian.Uint32(ipSrc.To4())
	dst := binary.BigEndian.Uint32(ipDst.To4())

	arr := make([]uint16, 6)

	arr[0] = uint16(src >> 16)
	arr[1] = uint16(src & 0xFFFF)
	arr[2] = uint16(dst >> 16)
	arr[3] = uint16(dst & 0xFFFF)
	arr[4] = uint16(proto)
	arr[5] = length

	sum := uint32(0)
	for i := 0; i < len(arr); i++ {
		sum += uint32(arr[i])
	}

	return sum
}

func IPv4AddrChecksum(hdr *ipv4.Header) uint32 {

	src := binary.BigEndian.Uint32(hdr.Src.To4())
	dst := binary.BigEndian.Uint32(hdr.Dst.To4())

	arr := make([]uint16, 4)

	arr[0] = uint16(src >> 16)
	arr[1] = uint16(src & 0xFFFF)
	arr[2] = uint16(dst >> 16)
	arr[3] = uint16(dst & 0xFFFF)

	sum := uint32(0)
	for i := 0; i < len(arr); i++ {
		sum += uint32(arr[i])
	}
	return sum
}

// IPv4UDPChecksum calculates UDP checksum for case if L3 protocol is IPv4.
func IPv4UDPChecksum(ip *ipv4.Header, udp *UDPHdr, data []byte) uint16 {

	dbug.Printf("dataLength=%d\n", len(data))

	arr := make([]uint16, 6)

	// Create pseudo-header for UDP/IPv4
	sum := PseudoHdrIPv4Checksum(ip.Src, ip.Dst, ip.Protocol, udp.Length)

	// UDP header
	arr[2] = uint16(udp.SrcPort)
	arr[3] = uint16(udp.DstPort)
	arr[4] = uint16(udp.Length)
	arr[5] = uint16(0)

	for i := 0; i < len(arr); i++ {
		sum += uint32(arr[i])
	}
	sum += dataChecksum(data, len(data))

	retSum := ^reduceChecksum(sum)
	dbug.Printf("sum %08x, retSum %04x\n", sum, retSum)

	// If the checksum calculation results in the value zero (all 16 bits 0) it
	// should be sent as the one's complement (all 1s).
	if retSum == 0 {
		retSum = ^retSum
	}
	dbug.Printf("retSum=%04x\n", retSum)
	return retSum
}

func TCPChecksum(tcp *TCPHdr) uint32 {
	return uint32(SwapUint16(tcp.SrcPort)) +
		uint32(SwapUint16(tcp.DstPort)) +
		uint32(SwapUint16(uint16(tcp.SeqNum>>16))) +
		uint32(SwapUint16(uint16(tcp.SeqNum))) +
		uint32(SwapUint16(uint16(tcp.AckNum>>16))) +
		uint32(SwapUint16(uint16(tcp.AckNum))) +
		(uint32(tcp.Flags)>>12)<<8 + // Data offset
		uint32(tcp.Flags&0xFF) +
		uint32(SwapUint16(tcp.Window)) +
		uint32(SwapUint16(tcp.Urgent))
}

// IPv4TCPChecksum calculates TCP checksum for case if L3
// protocol is IPv4. Here data pointer should point to end of minimal
// TCP header because we consider TCP options as part of data.
func IPv4TCPChecksum(ip *ipv4.Header, tcp *TCPHdr, data []byte) uint16 {
	dbug.Printf("dataLength=%d\n", len(data))

	arr := make([]uint16, 12)

	// Create pseudo-header for TCP/IPv4
	sum := PseudoHdrIPv4Checksum(ip.Src, ip.Dst, ip.Protocol, tcp.HdrLen+uint16(len(data)))

	// TCP header
	arr[2] = uint16(tcp.SrcPort)
	arr[3] = uint16(tcp.DstPort)
	arr[4] = uint16(tcp.SeqNum >> 16)
	arr[5] = uint16(tcp.SeqNum)
	arr[6] = uint16(tcp.AckNum >> 16)
	arr[7] = uint16(tcp.AckNum)
	flags := ((tcp.HdrLen >> 2) << 12) | tcp.Flags
	arr[8] = uint16(flags)
	arr[9] = uint16(tcp.Window)
	arr[10] = uint16(0) // Checksum
	arr[11] = uint16(tcp.Urgent)

	for i := 0; i < len(arr); i++ {
		sum += uint32(arr[i])
	}
	sum += dataChecksum(data, len(data))

	retSum := ^reduceChecksum(sum)
	dbug.Printf("sum %08x, retSum %04x\n", sum, retSum)

	// If the checksum calculation results in the value zero (all 16 bits 0) it
	// should be sent as the one's complement (all 1s).
	if retSum == 0 {
		retSum = ^retSum
	}
	dbug.Printf("retSum=%04x\n", retSum)
	return retSum
}

func IPv6AddrChecksum(ip *ipv6.Header) uint32 {

	src := make([]uint16, 8)
	src[0] = binary.BigEndian.Uint16(ip.Src.To16()[0:2])
	src[1] = binary.BigEndian.Uint16(ip.Src.To16()[2:4])
	src[2] = binary.BigEndian.Uint16(ip.Src.To16()[4:6])
	src[3] = binary.BigEndian.Uint16(ip.Src.To16()[6:8])
	src[4] = binary.BigEndian.Uint16(ip.Src.To16()[8:10])
	src[5] = binary.BigEndian.Uint16(ip.Src.To16()[10:12])
	src[6] = binary.BigEndian.Uint16(ip.Src.To16()[12:14])
	src[7] = binary.BigEndian.Uint16(ip.Src.To16()[14:16])

	dst := make([]uint16, 8)
	dst[0] = binary.BigEndian.Uint16(ip.Dst.To16()[0:2])
	dst[1] = binary.BigEndian.Uint16(ip.Dst.To16()[2:4])
	dst[2] = binary.BigEndian.Uint16(ip.Dst.To16()[4:6])
	dst[3] = binary.BigEndian.Uint16(ip.Dst.To16()[6:8])
	dst[4] = binary.BigEndian.Uint16(ip.Dst.To16()[8:10])
	dst[5] = binary.BigEndian.Uint16(ip.Dst.To16()[10:12])
	dst[6] = binary.BigEndian.Uint16(ip.Dst.To16()[12:14])
	dst[7] = binary.BigEndian.Uint16(ip.Dst.To16()[14:16])

	return uint32(src[0]) + uint32(src[1]) + uint32(src[2]) + uint32(src[3]) +
		uint32(src[4]) + uint32(src[5]) + uint32(src[6]) + uint32(src[7]) +
		uint32(dst[0]) + uint32(dst[1]) + uint32(dst[2]) + uint32(dst[3]) +
		uint32(dst[4]) + uint32(dst[5]) + uint32(dst[6]) + uint32(dst[7])
}

// IPv6UDPChecksum calculates UDP checksum for case if L3 protocol is IPv6.
func IPv6UDPChecksum(ip *ipv6.Header, udp *UDPHdr, data []byte) uint16 {
	dataLength := SwapUint16(uint16(ip.PayloadLen))

	sum := dataChecksum(data, int(dataLength-UDPDefaultLen))

	sum += IPv6AddrChecksum(ip) +
		uint32(SwapUint16(udp.Length)) +
		uint32(ip.NextHeader) +
		uint32(SwapUint16(udp.SrcPort)) +
		uint32(SwapUint16(udp.DstPort)) +
		uint32(SwapUint16(udp.Length))

	retSum := ^reduceChecksum(sum)
	// If the checksum calculation results in the value zero (all 16 bits 0) it
	// should be sent as the one's complement (all 1s).
	if retSum == 0 {
		retSum = ^retSum
	}
	return retSum
}

// IPv6TCPChecksum calculates TCP checksum for case if L3 protocol is IPv6.
func IPv6TCPChecksum(ip *ipv6.Header, tcp *TCPHdr, data []byte) uint16 {
	dataLength := SwapUint16(uint16(ip.PayloadLen))

	sum := dataChecksum(data, int(dataLength-TCPMinLen))

	sum += IPv6AddrChecksum(ip) +
		uint32(dataLength) +
		uint32(ip.NextHeader) +
		TCPChecksum(tcp)

	return ^reduceChecksum(sum)
}

// IPv4ICMPChecksum calculates ICMP checksum in case if L3
// protocol is IPv4.
func IPv4ICMPChecksum(hdr *ipv4.Header, icmp *ICMPHeader, data []byte) uint16 {
	dataLength := SwapUint16(uint16(hdr.TotalLen)) - IPv4MinLen - ICMPv4MinLen

	sum := uint32(uint16(icmp.Type)<<8|uint16(icmp.Code)) +
		uint32(SwapUint16(icmp.Identifier)) +
		uint32(SwapUint16(icmp.SeqNum)) +
		dataChecksum(data, int(dataLength))

	return ^reduceChecksum(sum)
}

// IPv6ICMPChecksum calculates ICMP checksum in case if L3 protocol is IPv6.
func IPv6ICMPChecksum(ip *ipv6.Header, icmp *ICMPHeader, data []byte) uint16 {
	dataLength := SwapUint16(uint16(ip.PayloadLen))

	// ICMP payload
	sum := dataChecksum(data, int(dataLength-ICMPv6MinLen))

	sum += IPv6AddrChecksum(ip) + // IPv6 Header
		uint32(dataLength) +
		uint32(ip.NextHeader) +
		// ICMP header excluding checksum
		uint32(uint16(icmp.Type)<<8|uint16(icmp.Code)) +
		uint32(SwapUint16(icmp.Identifier)) +
		uint32(SwapUint16(icmp.SeqNum))

	return ^reduceChecksum(sum)
}

// Software calculation of protocol headers. It is required for hardware checksum
// calculation offload

// PseudoHdrIPv4TCPChecksum implements one step of TCP checksum calculation.
// Separately computes checksum for TCP pseudo-header for case if L3 protocol is IPv4.
// This precalculation is required for checksum compute by hardware offload.
// Result should be put into TCP.Cksum field.
func PseudoHdrIPv4TCPChecksum(hdr *ipv4.Header) uint16 {
	dataLength := SwapUint16(uint16(hdr.TotalLen)) - IPv4MinLen
	pHdrCksum := IPv4AddrChecksum(hdr) +
		uint32(hdr.Protocol) +
		uint32(dataLength)
	return reduceChecksum(pHdrCksum)
}

// PseudoHdrIPv4UDPChecksum implements one step of UDP checksum calculation. Separately computes checksum
// for UDP pseudo-header for case if L3 protocol is IPv4.
// This precalculation is required for checksum compute by hardware offload.
// Result should be put into UDP.DgramCksum field.
func PseudoHdrIPv4UDPChecksum(hdr *ipv4.Header, udp *UDPHdr) uint16 {
	pHdrCksum := IPv4AddrChecksum(hdr) +
		uint32(hdr.Protocol) +
		uint32(SwapUint16(udp.Length))
	return reduceChecksum(pHdrCksum)
}

// PseudoHdrIPv6TCPChecksum implements one step of TCP checksum calculation. Separately computes checksum
// for TCP pseudo-header for case if L3 protocol is IPv6.
// This precalculation is required for checksum compute by hardware offload.
// Result should be put into TCP.Cksum field.
func PseudoHdrIPv6TCPChecksum(hdr *ipv6.Header) uint16 {
	dataLength := SwapUint16(uint16(hdr.PayloadLen))
	pHdrCksum := IPv6AddrChecksum(hdr) +
		uint32(dataLength) +
		uint32(hdr.NextHeader)
	return reduceChecksum(pHdrCksum)
}

// PseudoHdrIPv6UDPChecksum implements one step of UDP checksum calculation. Separately computes checksum
// for UDP pseudo-header for case if L3 protocol is IPv6.
// This precalculation is required for checksum compute by hardware offload.
// Result should be put into UDP.DgramCksum field.
func PseudoHdrIPv6UDPChecksum(hdr *ipv6.Header, udp *UDPHdr) uint16 {
	pHdrCksum := IPv6AddrChecksum(hdr) +
		uint32(hdr.NextHeader) +
		uint32(SwapUint16(udp.Length))
	return reduceChecksum(pHdrCksum)
}
